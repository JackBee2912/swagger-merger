# Swagger Merger
cd
A powerful CLI tool and Go library for merging multiple Swagger/OpenAPI files into a single, unified specification.

## 🚀 Features

- **Multi-format Support**: Merge Swagger 2.0 and OpenAPI 3.0 files
- **Flexible Input**: Support for individual files, directories, and URL-based swagger files
- **Custom Server Configuration**: Override server URLs and descriptions
- **Cross-platform**: Available for Linux, macOS, and Windows
- **Docker Support**: Containerized for easy CI/CD integration
- **Statistics**: Get detailed statistics about merged files
- **Verbose Output**: Detailed logging for debugging

## 📦 Installation

### Option 1: Download Pre-built Binary

Download the latest release for your platform from the [Releases page](https://github.com/JackBee2912/swagger-merger/releases).

```bash
# Linux
wget https://github.com/JackBee2912/swagger-merger/releases/latest/download/swagger-merger-linux-amd64
chmod +x swagger-merger-linux-amd64
sudo mv swagger-merger-linux-amd64 /usr/local/bin/swagger-merger

# macOS
wget https://github.com/JackBee2912/swagger-merger/releases/latest/download/swagger-merger-darwin-amd64
chmod +x swagger-merger-darwin-amd64
sudo mv swagger-merger-darwin-amd64 /usr/local/bin/swagger-merger

# Windows
# Download swagger-merger-windows-amd64.exe and add to PATH
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/JackBee2912/swagger-merger.git
cd swagger-merger

# Build the CLI tool
make build-cli

# Install globally (optional)
make install
```

### Option 3: Using Go

```bash
go install github.com/JackBee2912/swagger-merger/cmd/swagger-merger@latest
```

### Option 4: Using Docker

```bash
# Pull the image
docker pull JackBee2912/swagger-merger:latest

# Run the tool
docker run --rm -v $(pwd):/workspace JackBee2912/swagger-merger:latest --help
```

## 🎯 Quick Start

### Basic Usage

```bash
# Merge specific files
swagger-merger --input file1.yaml,file2.yaml --output merged.yaml

# Merge files from directory
swagger-merger --input ./docs --output merged.yaml

# Merge with custom servers
swagger-merger --input ./docs --output merged.yaml \
  --servers "https://api-dev.com:Development,https://api.com:Production"

# Verbose output with statistics
swagger-merger --input ./docs --output merged.yaml --verbose --stats
```

### Advanced Usage

```bash
# Merge from multiple sources
swagger-merger --input "./docs,./api-specs,https://api.example.com/swagger.json" \
  --output combined.yaml \
  --servers "https://api-dev.example.com:Development,https://api.example.com:Production" \
  --verbose --stats

# Use custom file pattern
swagger-merger --input ./docs --pattern "*.swagger.yaml" --output merged.yaml

# Merge with default servers
swagger-merger --input ./docs --output merged.yaml
```

## 📖 CLI Reference

### Commands

```bash
swagger-merger [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--input` | string | | **Required**. Comma-separated list of input swagger files or directories |
| `--output` | string | `merged_swagger.yaml` | Output file path |
| `--pattern` | string | `*.{yaml,yml}` | File pattern for directory scanning |
| `--servers` | string | | Comma-separated list of server URLs (format: `url:description`) |
| `--verbose` | bool | `false` | Enable verbose output |
| `--stats` | bool | `false` | Show statistics after merging |
| `--version` | bool | `false` | Show version information |
| `--help` | bool | `false` | Show help message |

### Server Format

The `--servers` flag accepts URLs in the following format:

```
url:description
```

Examples:
- `https://api-dev.com:Development`
- `https://api.com:Production`
- `http://localhost:8080:Local Development`

Multiple servers can be specified by separating them with commas:

```bash
--servers "https://api-dev.com:Development,https://api.com:Production,http://localhost:8080:Local"
```

### Default Servers

If no servers are specified, the tool uses these default servers:

- `https://api-dev.domain.com` (Development Environment)
- `https://api-test.domain.com` (Test Environment)
- `https://api-stg.domain.com` (Staging Environment)
- `https://api.domain.com` (Production Environment)

## 🔧 Library Usage

### Import

```go
import "github.com/JackBee2912/swagger-merger/pkg/merger"
```

### Basic Example

```go
package main

import (
    "log"
    "github.com/JackBee2912/swagger-merger/pkg/merger"
)

func main() {
    config := merger.Config{
        InputPaths: []string{
            "file1.yaml",
            "file2.yaml",
        },
        OutputPath: "merged.yaml",
        Servers: []merger.Server{
            {URL: "https://api-dev.com", Description: "Development"},
            {URL: "https://api.com", Description: "Production"},
        },
    }

    mergerInstance := merger.New(config)
    
    if err := mergerInstance.Merge(); err != nil {
        log.Fatalf("Error merging files: %v", err)
    }

    // Get statistics
    stats, err := mergerInstance.GetStats()
    if err != nil {
        log.Printf("Error getting stats: %v", err)
    } else {
        log.Printf("Statistics: %+v", stats)
    }
}
```

### Advanced Example

```go
package main

import (
    "log"
    "github.com/JackBee2912/swagger-merger/pkg/merger"
)

func main() {
    // Use default servers
    config := merger.Config{
        OutputPath: "merged.yaml",
    }

    mergerInstance := merger.New(config)

    // Merge from directory
    if err := mergerInstance.MergeFromDirectory("./docs", "*.{yaml,yml}"); err != nil {
        log.Fatalf("Error merging from directory: %v", err)
    }
}
```

## 🐳 Docker Usage

### Basic Docker Command

```bash
docker run --rm -v $(pwd):/workspace JackBee2912/swagger-merger:latest \
  --input /workspace/docs \
  --output /workspace/merged.yaml
```

### Docker with Custom Servers

```bash
docker run --rm -v $(pwd):/workspace JackBee2912/swagger-merger:latest \
  --input /workspace/docs \
  --output /workspace/merged.yaml \
  --servers "https://api-dev.com:Development,https://api.com:Production" \
  --verbose --stats
```

### Docker Compose Example

```yaml
version: '3.8'
services:
  swagger-merger:
    image: JackBee2912/swagger-merger:latest
    volumes:
      - ./docs:/workspace/docs
      - ./output:/workspace/output
    command: >
      --input /workspace/docs
      --output /workspace/output/merged.yaml
      --servers "https://api-dev.com:Development,https://api.com:Production"
      --verbose --stats
```

## 🔄 CI/CD Integration

### GitHub Actions

```yaml
name: Merge Swagger Files

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  merge-swagger:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Merge Swagger Files
      run: |
        wget https://github.com/JackBee2912/swagger-merger/releases/latest/download/swagger-merger-linux-amd64
        chmod +x swagger-merger-linux-amd64
        ./swagger-merger-linux-amd64 \
          --input ./docs \
          --output ./merged-swagger.yaml \
          --servers "https://api-dev.com:Development,https://api.com:Production" \
          --verbose --stats
    
    - name: Upload merged file
      uses: actions/upload-artifact@v4
      with:
        name: merged-swagger
        path: merged-swagger.yaml
```

### GitLab CI

```yaml
merge_swagger:
  image: JackBee2912/swagger-merger:latest
  script:
    - swagger-merger \
        --input /workspace/docs \
        --output /workspace/merged-swagger.yaml \
        --servers "https://api-dev.com:Development,https://api.com:Production" \
        --verbose --stats
  artifacts:
    paths:
      - merged-swagger.yaml
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    
    stages {
        stage('Merge Swagger') {
            steps {
                sh '''
                    wget https://github.com/JackBee2912/swagger-merger/releases/latest/download/swagger-merger-linux-amd64
                    chmod +x swagger-merger-linux-amd64
                    ./swagger-merger-linux-amd64 \\
                        --input ./docs \\
                        --output ./merged-swagger.yaml \\
                        --servers "https://api-dev.com:Development,https://api.com:Production" \\
                        --verbose --stats
                '''
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: 'merged-swagger.yaml', fingerprint: true
        }
    }
}
```

## 📊 Output Statistics

When using the `--stats` flag, the tool provides detailed statistics:

```
📊 Statistics:
  Total files: 3
  Total paths: 15
  Total schemas: 25
  Total tags: 8
```

## 🔍 Troubleshooting

### Common Issues

1. **"No valid input files found"**
   - Check that the input paths are correct
   - Ensure files have `.yaml` or `.yml` extensions
   - Verify file permissions

2. **"Failed to detect version"**
   - Ensure files are valid Swagger 2.0 or OpenAPI 3.0
   - Check for syntax errors in YAML/JSON

3. **"HTTP request failed"**
   - Verify URL accessibility
   - Check network connectivity
   - Ensure URLs return valid swagger files

4. **Server URLs not applied**
   - Check server format: `url:description`
   - Ensure no extra spaces around colons
   - Verify URLs are properly formatted

### Debug Mode

Use the `--verbose` flag for detailed logging:

```bash
swagger-merger --input ./docs --output merged.yaml --verbose
```

This will show:
- File discovery process
- Version detection results
- Merge progress
- Server configuration details

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/JackBee2912/swagger-merger/issues)
- **Discussions**: [GitHub Discussions](https://github.com/JackBee2912/swagger-merger/discussions)
- **Documentation**: [Wiki](https://github.com/JackBee2912/swagger-merger/wiki)

## 🙏 Acknowledgments

- Built with [kin-openapi](https://github.com/getkin/kin-openapi)
- Inspired by the need for unified API documentation
- Thanks to all contributors and users 