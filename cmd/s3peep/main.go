package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lausser/s3peep/internal/config"
	"github.com/lausser/s3peep/internal/handlers"
	"github.com/lausser/s3peep/internal/s3"
)

var (
	configFlag = flag.String("config", "", "Path to config file (or set $CONFIG)")
)

func getConfigPath() string {
	if envPath := os.Getenv("CONFIG"); envPath != "" {
		return envPath
	}
	if *configFlag != "" {
		return *configFlag
	}
	return config.DefaultConfigPath()
}

func main() {
	flag.CommandLine.Parse(os.Args[1:])
	cfgPath := getConfigPath()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		return
	}

	switch args[0] {
	case "profile":
		cli := config.NewCLI(cfgPath)
		if err := cli.Run(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "serve":
		port, err := parseServeArgs(args[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		runServer(cfgPath, port)
	case "init":
		if err := config.CreateDefaultConfig(cfgPath); err != nil {
			log.Fatalf("Failed to create config: %v", err)
		}
		fmt.Printf("Created config at %s\n", cfgPath)
		config.PrintConfigExample()
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  s3peep init                              - Create default config file")
	fmt.Println("  s3peep profile add --name NAME --region REGION --access-key KEY --secret SECRET [--bucket BUCKET] [--endpoint URL]")
	fmt.Println("  s3peep profile list                      - List profiles")
	fmt.Println("  s3peep profile switch --name NAME        - Switch active profile")
	fmt.Println("  s3peep profile remove --name NAME        - Remove a profile")
	fmt.Println("  s3peep serve --port 8080                 - Start web server")
	fmt.Println("")
	fmt.Println("Configuration:")
	fmt.Println("  --config FILE   Config file path (default: ~/.config/s3peep/config.json)")
	fmt.Println("  $CONFIG         Environment variable for config file path")
}

func parseServeArgs(args []string) (int, error) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	port := fs.Int("port", 8080, "HTTP server port")
	if err := fs.Parse(args); err != nil {
		return 0, err
	}
	if fs.NArg() > 0 {
		return 0, fmt.Errorf("unknown command: %s", fs.Arg(0))
	}
	return *port, nil
}

func runServer(cfgPath string, port int) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	profile := config.GetActiveProfile(cfg)
	if profile == nil {
		fmt.Println("No active profile. Use 's3peep profile add' to create one.")
		os.Exit(1)
	}

	s3Client, err := s3.NewClient(profile, profile.Bucket)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	ctx := context.Background()
	if err := s3Client.TestConnection(ctx); err != nil {
		log.Fatalf("Failed to connect to S3: %v", err)
	}

	if profile.Bucket != "" {
		fmt.Printf("Connected to S3 bucket: %s\n", profile.Bucket)
	} else {
		fmt.Println("Connected to S3 (no bucket selected)")
	}

	handler := handlers.NewAPIHandler(cfg, cfgPath, s3Client)
	http.HandleFunc("/", handler.Handle)

	fmt.Printf("Starting server on port %d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
