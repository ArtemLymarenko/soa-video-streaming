package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2asrv"
	"github.com/a2aproject/a2a-go/a2asrv/eventqueue"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"

	"soa-video-streaming/services/a2a-server/internal/tools"
	"soa-video-streaming/services/ai-assistant/pkg/gemini"
)

type ProducerAgent struct {
	tools        *tools.ContentTools
	geminiClient *gemini.Client
}

var _ a2asrv.AgentExecutor = (*ProducerAgent)(nil)

func NewProducerAgent(tools *tools.ContentTools, geminiClient *gemini.Client) a2asrv.AgentExecutor {
	return &ProducerAgent{
		tools:        tools,
		geminiClient: geminiClient,
	}
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewProducerAgent),
	)
}

func (a *ProducerAgent) Execute(ctx context.Context, req *a2asrv.RequestContext, queue eventqueue.Queue) error {
	userPrompt := req.Message.Parts[0].(a2a.TextPart).Text
	if userPrompt == "" {
		return fmt.Errorf("empty prompt in request")
	}

	logrus.WithFields(logrus.Fields{
		"task_id": req.TaskID,
		"prompt":  userPrompt,
	}).Info("Executing ProducerAgent")

	processor := gemini.NewMessageProcessor(a.geminiClient, "gemini-2.5-flash-lite")
	a.registerHandlers(processor)

	geminiTools := a.getGeminiTools()
	modelConfig := &gemini.ModelConfig{
		Name:  "gemini-2.5-flash-lite",
		Tools: geminiTools,
	}

	_ = a.geminiClient.RegisterModel(modelConfig)

	resp, llmErr := processor.ProcessWithToolLoop(ctx, userPrompt, 5)

	var msg *a2a.Message
	if llmErr == nil {
		msg = &a2a.Message{
			Role: a2a.MessageRoleAgent,
			Parts: []a2a.Part{
				a2a.TextPart{Text: resp.Text},
			},
		}
	} else {
		logrus.WithError(llmErr).Error("Gemini processing failed")
	}

	event := a2a.NewStatusUpdateEvent(req, "success", msg)
	if err := queue.Write(ctx, event); err != nil {
		logrus.WithError(err).Error("Failed to write response to queue")
		return err
	}

	return nil
}

func (a *ProducerAgent) Cancel(ctx context.Context, req *a2asrv.RequestContext, queue eventqueue.Queue) error {
	logrus.Infof("Task %s cancelled", req.TaskID)
	return nil
}

func (a *ProducerAgent) getGeminiTools() []*gemini.Tool {
	return []*gemini.Tool{
		gemini.NewTool("CheckCategory", "Check if a movie category exists by name").
			AddParameter("name", "string", "Name of the category to check").
			MarkRequired("name"),
		gemini.NewTool("CreateCategory", "Create a new movie category").
			AddParameter("name", "string", "Name of the category to create").
			AddParameter("description", "string", "Description of the category").
			MarkRequired("name"),
		gemini.NewTool("CreateMovie", "Create a new movie or series").
			AddParameter("title", "string", "Title of the movie/series").
			AddParameter("description", "string", "Description of the content").
			AddParameter("category_name", "string", "Name of the category (must exist)").
			AddEnumParameter("type", "Type of content (movie or series)", "movie", "series").
			AddParameter("duration", "integer", "Duration in minutes").
			MarkRequired("title", "category_name", "type"),
	}
}

func (a *ProducerAgent) registerHandlers(p *gemini.MessageProcessor) {
	p.RegisterHandler("CheckCategory", func(ctx context.Context, _ string, args json.RawMessage) (string, error) {
		var input tools.CheckCategoryInput
		if err := json.Unmarshal(args, &input); err != nil {
			return "", err
		}

		res, err := a.tools.CheckCategory(ctx, input)
		return a.marshalToolResult(res, err)
	})

	p.RegisterHandler("CreateCategory", func(ctx context.Context, _ string, args json.RawMessage) (string, error) {
		var input tools.CreateCategoryInput
		if err := json.Unmarshal(args, &input); err != nil {
			return "", err
		}

		res, err := a.tools.CreateCategory(ctx, input)
		return a.marshalToolResult(res, err)
	})

	p.RegisterHandler("CreateMovie", func(ctx context.Context, _ string, args json.RawMessage) (string, error) {
		var input tools.CreateMovieInput
		if err := json.Unmarshal(args, &input); err != nil {
			return "", err
		}
		res, err := a.tools.CreateMovie(ctx, input)
		return a.marshalToolResult(res, err)
	})
}

func (a *ProducerAgent) marshalToolResult(res interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}

	b, _ := json.Marshal(res)
	return string(b), nil
}
