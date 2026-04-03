package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/lausser/s3peep/internal/config"
	"github.com/lausser/s3peep/internal/handlers"
	"github.com/lausser/s3peep/internal/s3"
)

var (
	configFlag *string
	debugFlag  *bool
)

func init() {
	// Create a custom flag set for consistent double-dash output
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	// Discard default output - we'll handle errors ourselves
	fs.SetOutput(io.Discard)
	
	// Override usage to show double dashes
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "  --%s %s\n", f.Name, f.DefValue)
			fmt.Fprintf(os.Stderr, "\t%s\n", f.Usage)
		})
	}
	
	// Define flags on custom flag set
	configFlag = fs.String("config", "", "Path to config file (or set $CONFIG)")
	debugFlag = fs.Bool("debug", false, "Enable debug logging")
	
	// Replace the default flag.CommandLine with our custom one
	flag.CommandLine = fs
}

// normalizeFlagError converts single-dash error messages to double-dash format
func normalizeFlagError(err error) string {
	errMsg := err.Error()
	// Replace single-dash with double-dash in error messages
	// Replace " -flag" with " --flag" at end of message or before end quote
	re := regexp.MustCompile(`([:\s"])-(\w+)`)
	errMsg = re.ReplaceAllString(errMsg, "${1}--${2}")
	return errMsg
}

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
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		// Convert single-dash error messages to double-dash
		errMsg := normalizeFlagError(err)
		// Don't print "help requested" as an error - just show usage
		if !strings.Contains(errMsg, "help requested") {
			fmt.Fprintf(os.Stderr, "%s\n", errMsg)
		}
		flag.Usage()
		if strings.Contains(errMsg, "help requested") {
			return
		}
		os.Exit(1)
	}
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
			if err.Error() != "" {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}
	case "serve":
		port, debug, err := parseServeArgs(args[1:])
		if err != nil {
			if err.Error() != "" {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}
		runServer(cfgPath, port, debug)
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
	fmt.Println("  s3peep serve --port 8080 [--debug]       - Start web server")
	fmt.Println("")
	fmt.Println("Configuration:")
	fmt.Println("  --config FILE   Config file path (default: ~/.config/s3peep/config.json)")
	fmt.Println("  $CONFIG         Environment variable for config file path")
	fmt.Println("  --debug         Enable debug logging for API requests")
}

func parseServeArgs(args []string) (int, bool, error) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	// Override usage to show double dashes
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of serve:\n")
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "  --%s %s\n", f.Name, f.DefValue)
			fmt.Fprintf(os.Stderr, "\t%s\n", f.Usage)
		})
	}
	port := fs.Int("port", 8080, "HTTP server port")
	debug := fs.Bool("debug", false, "Enable debug logging")
	if err := fs.Parse(args); err != nil {
		// Normalize error and check for help request
		errMsg := normalizeFlagError(err)
		if !strings.Contains(errMsg, "help requested") {
			fmt.Fprintf(os.Stderr, "%s\n", errMsg)
			fs.Usage()
			return 0, false, fmt.Errorf("")
		}
		// For help request, Usage was already called by flag package
		return 0, false, fmt.Errorf("")
	}
	if fs.NArg() > 0 {
		return 0, false, fmt.Errorf("unknown command: %s", fs.Arg(0))
	}
	return *port, *debug, nil
}

// generateToken creates a cryptographically secure random token
func generateToken() (string, error) {
	// Generate 32 bytes of random data (256 bits of entropy)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Encode to base64 URL-safe string
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func runServer(cfgPath string, port int, debug bool) {
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

	// Generate secure token
	token, err := generateToken()
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	if profile.Bucket != "" {
		fmt.Printf("Connected to S3 bucket: %s\n", profile.Bucket)
	} else {
		fmt.Println("Connected to S3 (no bucket selected)")
	}

	// Create handler with token and debug mode
	handler := handlers.NewAPIHandler(cfg, cfgPath, s3Client, token, debug)
	
	// Use the handler with token-based routing
	http.HandleFunc("/", handler.Handle)

	// Print access URL with token
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                    S3 File Browser Ready                      ║\n")
	fmt.Printf("╠════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Access URL: http://localhost:%d/%s  ║\n", port, token)
	fmt.Printf("╚════════════════════════════════════════════════════════════════╝\n\n")
	fmt.Printf("Starting server on port %d...\n", port)
	if debug {
		fmt.Println("Debug mode enabled - API requests will be logged")
	}
	fmt.Println("Press Ctrl+C to stop the server\n")
	
	if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
