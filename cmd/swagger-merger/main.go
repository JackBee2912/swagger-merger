package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"swagger-merger/pkg/merger"
)

func main() {
	var (
		inputPaths = flag.String("input", "", "Comma-separated list of input swagger files or directories")
		outputPath = flag.String("output", "merged_swagger.yaml", "Output file path")
		pattern    = flag.String("pattern", "*.yaml", "File pattern for directory scanning (supports comma-separated patterns)")
		servers    = flag.String("servers", "", "Comma-separated list of server URLs (format: url:description)")
		version    = flag.Bool("version", false, "Show version information")
		help       = flag.Bool("help", false, "Show help information")
		verbose    = flag.Bool("verbose", false, "Enable verbose output")
		stats      = flag.Bool("stats", false, "Show statistics after merging")
	)

	flag.Parse()

	// Show version
	if *version {
		fmt.Println("swagger-merger v1.0.0")
		fmt.Println("A tool for merging multiple Swagger/OpenAPI files")
		return
	}

	// Show help
	if *help {
		showHelp()
		return
	}

	// Validate required flags
	if *inputPaths == "" {
		log.Fatal("❌ Error: --input flag is required")
	}

	if *outputPath == "" {
		log.Fatal("❌ Error: --output flag is required")
	}

	// Parse servers
	var serverConfigs []merger.Server
	if *servers != "" {
		serverList := strings.Split(*servers, ",")
		for _, server := range serverList {
			server = strings.TrimSpace(server)
			if server == "" {
				continue
			}

			// Find the last colon to separate URL from description
			lastColonIndex := strings.LastIndex(server, ":")
			if lastColonIndex > 0 && lastColonIndex < len(server)-1 {
				// Check if it's not part of a protocol (http://, https://)
				beforeColon := server[:lastColonIndex]
				if !strings.HasSuffix(beforeColon, "//") {
					serverConfigs = append(serverConfigs, merger.Server{
						URL:         strings.TrimSpace(server[:lastColonIndex]),
						Description: strings.TrimSpace(server[lastColonIndex+1:]),
					})
				} else {
					// No description provided, use default
					serverConfigs = append(serverConfigs, merger.Server{
						URL:         server,
						Description: "API Server",
					})
				}
			} else {
				// No colon found or colon at the end, treat as URL only
				serverConfigs = append(serverConfigs, merger.Server{
					URL:         server,
					Description: "API Server",
				})
			}
		}
	}

	// Use default servers if none provided
	if len(serverConfigs) == 0 {
		serverConfigs = merger.DefaultServers()
	}

	// Create merger config
	config := merger.Config{
		OutputPath: *outputPath,
		Servers:    serverConfigs,
	}

	// Create merger instance
	mergerInstance := merger.New(config)

	// Parse input paths
	inputPathList := strings.Split(*inputPaths, ",")
	var allInputPaths []string

	// Parse patterns
	patterns := strings.Split(*pattern, ",")
	for i, p := range patterns {
		patterns[i] = strings.TrimSpace(p)
	}

	for _, inputPath := range inputPathList {
		inputPath = strings.TrimSpace(inputPath)
		if inputPath == "" {
			continue
		}

		// Check if it's a directory
		info, err := os.Stat(inputPath)
		if err != nil {
			log.Printf("⚠️  Warning: Cannot access %s: %v", inputPath, err)
			continue
		}

		if info.IsDir() {
			// Scan directory for swagger files
			if *verbose {
				fmt.Printf("📁 Scanning directory: %s\n", inputPath)
			}

			err := filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() {
					// Check if file matches any of the patterns
					matched := false
					for _, pattern := range patterns {
						patternMatched, err := filepath.Match(pattern, filepath.Base(path))
						if err != nil {
							continue
						}
						if patternMatched {
							matched = true
							break
						}
					}

					if matched {
						allInputPaths = append(allInputPaths, path)
						if *verbose {
							fmt.Printf("  📄 Found: %s\n", path)
						}
					}
				}
				return nil
			})

			if err != nil {
				log.Printf("⚠️  Warning: Error scanning directory %s: %v", inputPath, err)
			}
		} else {
			// Single file
			allInputPaths = append(allInputPaths, inputPath)
			if *verbose {
				fmt.Printf("📄 Input file: %s\n", inputPath)
			}
		}
	}

	if len(allInputPaths) == 0 {
		log.Fatal("❌ Error: No valid input files found")
	}

	// Update config with found files
	config.InputPaths = allInputPaths
	mergerInstance = merger.New(config)

	// Perform merge
	if *verbose {
		fmt.Printf("🔄 Merging %d files...\n", len(allInputPaths))
	}

	if err := mergerInstance.Merge(); err != nil {
		log.Fatalf("❌ Error merging files: %v", err)
	}

	fmt.Printf("✅ Successfully merged %d files to: %s\n", len(allInputPaths), *outputPath)

	// Show statistics if requested
	if *stats {
		stats, err := mergerInstance.GetStats()
		if err != nil {
			log.Printf("⚠️  Warning: Could not get statistics: %v", err)
		} else {
			fmt.Println("📊 Statistics:")
			fmt.Printf("  Total files: %d\n", stats["total_files"])
			fmt.Printf("  Total paths: %d\n", stats["total_paths"])
			fmt.Printf("  Total schemas: %d\n", stats["total_schemas"])
			fmt.Printf("  Total tags: %d\n", stats["total_tags"])
		}
	}

	// Show server information
	if *verbose {
		fmt.Println("🌐 Configured servers:")
		for i, server := range serverConfigs {
			fmt.Printf("  %d. %s (%s)\n", i+1, server.URL, server.Description)
		}
	}
}

func showHelp() {
	fmt.Println("swagger-merger - A tool for merging multiple Swagger/OpenAPI files")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  swagger-merger [flags]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --input string     Comma-separated list of input swagger files or directories")
	fmt.Println("  --output string    Output file path (default: merged_swagger.yaml)")
	fmt.Println("  --pattern string   File pattern for directory scanning (default: *.yaml, supports comma-separated patterns)")
	fmt.Println("  --servers string   Comma-separated list of server URLs (format: url:description)")
	fmt.Println("  --version          Show version information")
	fmt.Println("  --help             Show this help message")
	fmt.Println("  --verbose          Enable verbose output")
	fmt.Println("  --stats            Show statistics after merging")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Merge specific files")
	fmt.Println("  swagger-merger --input file1.yaml,file2.yaml --output merged.yaml")
	fmt.Println("")
	fmt.Println("  # Merge files from directory")
	fmt.Println("  swagger-merger --input ./docs --output merged.yaml")
	fmt.Println("")
	fmt.Println("  # Merge with custom pattern")
	fmt.Println("  swagger-merger --input ./docs --output merged.yaml --pattern '*.yaml,*.yml'")
	fmt.Println("")
	fmt.Println("  # Merge with custom servers")
	fmt.Println("  swagger-merger --input ./docs --output merged.yaml --servers 'https://api-dev.com:Development,https://api.com:Production'")
	fmt.Println("")
	fmt.Println("  # Verbose output with statistics")
	fmt.Println("  swagger-merger --input ./docs --output merged.yaml --verbose --stats")
}
