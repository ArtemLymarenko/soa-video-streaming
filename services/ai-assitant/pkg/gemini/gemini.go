// Package gemini provides integration with Google Gemini AI API
// with support for function calling (Tools), declarative model configuration,
// and automatic message processing loops.
//
// Quick Start:
//
//	// Create client
//	client, err := gemini.NewClient(ctx, "your-api-key")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Register model
//	config := gemini.NewModelConfig("gemini-pro").
//		WithSystemInstruction("You are a helpful assistant")
//	client.RegisterModel(config)
//
//	// Create processor
//	processor := gemini.NewMessageProcessor(client, "gemini-pro")
//
//	// Send message
//	response, err := processor.SendMessage(ctx, "Hello!")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(response.Text)
//
// For detailed usage and examples, see README.md and examples_test.go
package gemini
