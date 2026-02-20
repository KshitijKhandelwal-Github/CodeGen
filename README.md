# CodeGen


![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Language](https://img.shields.io/badge/language-Go-green.svg)
![Version](https://img.shields.io/badge/version-0.1.0-orange.svg)
![Build](https://img.shields.io/badge/build-python-yellow.svg)



## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
- [Usage](#usage)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Development](#development)

- [License](#license)


## 🎯 Overview
CodeGen Pro is a high-performance command-line utility written in Go designed to automate and streamline the software development lifecycle through robust code generation and project scaffolding. Built on the Cobra command-line framework, the tool provides a structured, professional-grade interface for developers to manage repetitive coding tasks and standardize project architectures. While the project integrates the Gin web framework for potential service-based interactions, its primary footprint is a powerful CLI that bridges the gap between manual boilerplate creation and automated, configuration-driven development.

The codebase implements several specific features designed for developer productivity and operational reliability. It features a nested CLI architecture using the Cobra library, allowing for complex command structures and sub-commands. Configuration management is handled via the Viper ecosystem, providing native support for multiple formats including TOML, YAML, and environment variables. To enhance the developer experience, the tool incorporates real-time visual feedback through interactive progress bars and color-coded terminal output. Furthermore, the project includes a comprehensive cross-platform build system via a dedicated Makefile, enabling the generation of optimized binaries for Linux, macOS (both Intel and ARM), and Windows environments.

From a technical perspective, CodeGen Pro follows Go best practices by utilizing a multi-stage Docker build process to ensure minimal image sizes and secure, reproducible deployments. The architecture separates concerns by isolating the CLI entry points in the cmd directory from the core logic in internal and pkg folders, facilitating long-term maintainability. It utilizes Go 1.23 features and emphasizes performance through the use of ldflags to strip debug symbols during the build process. This tool is an ideal solution for platform engineers and backend developers who need to enforce architectural standards across a microservices ecosystem or rapidly scaffold new API components. By automating the "plumbing" of new services, CodeGen Pro allows engineering teams to maintain consistency and focus their efforts on core business logic rather than repetitive configuration.



## ✨ Features

- 🌐 Modern web application built with Gin
- 🚀 High-performance server-side rendering
- 🔒 Built-in security best practices
- 📱 Responsive design support



## 🛠️ Tech Stack

### Languages

- **Go** - 100.0% (19 files, 4699 lines of code)



### Framework
- **Gin** v1.9.1 - web framework


### Key Dependencies


- **github.com/fatih/color** `v1.18.0` - Go modules

- **github.com/pelletier/go-toml/v2** `v2.2.4` - Go modules

- **github.com/schollz/progressbar/v3** `v3.19.0` - Go modules

- **github.com/spf13/cobra** `v1.10.2` - Go modules

- **github.com/spf13/viper** `v1.21.0` - Go modules

- **github.com/stretchr/testify** `v1.11.1` - Go modules

- **gopkg.in/yaml.v3** `v3.0.1` - Go modules

- **github.com/gin-gonic/gin** `v1.9.1` - Go modules

- **github.com/spf13/cobra** `v1.8.0` - Go modules

- **github.com/spf13/viper** `v1.18.2` - Go modules

- **mongoose** `^7.0.0` - npm

- **express** `^4.18.2` - npm

- **dotenv** `^16.0.3` - npm

- **flask** `==2.3.0` - PyPI

- **requests** `>=2.28.0` - PyPI



### Build Tools
- **Package Manager**: pip
- **Build Tool**: python


## 🚀 Getting Started

### Prerequisites


- **Go** 1.21 or higher
- **Git** for version control


### Installation

### Using homebrew (Recommended for MacOS)
```bash
brew tap KshitijKhandelwal-Github/tap
brew install codegen
```

### Quick Install (Recommended)

**macOS & Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/codegen-pro/main/install.sh | bash
```

**Or download manually:**
- [macOS (Intel)](https://github.com/YOUR_USERNAME/codegen-pro/releases/latest/download/codegen-macos-amd64)
- [macOS (Apple Silicon)](https://github.com/YOUR_USERNAME/codegen-pro/releases/latest/download/codegen-macos-arm64)
- [Linux (amd64)](https://github.com/YOUR_USERNAME/codegen-pro/releases/latest/download/codegen-linux-amd64)
- [Windows (amd64)](https://github.com/YOUR_USERNAME/codegen-pro/releases/latest/download/codegen-windows-amd64.exe)

### Verify Installation
```bash
codegen --version
```

### Usage
```bash
# Works from any directory!
cd ~/my-project
codegen init
codegen generate --readme --docker
```

## 🔄 Updating
```bash
# Re-run install script
curl -fsSL https://raw.githubusercontent.com/KshitijKhandelwal-Github/CodeGen/main/install.sh | bash
```

## 🗑️ Uninstalling

### For Homebrew
```bash
brew uninstall codegen
brew untap KshitijKhandelwal-Github/tap
```

### Other methods
```bash
sudo rm /usr/local/bin/codegen
# or
rm ~/.local/bin/codegen
```



## 📖 Basic Usage

### Initializing the configuration for the project
```bash
cd ~/any-project
codegen init
```

### Check CodeGen configuration for the project
```bash
codegen config show
```

### Generate files
```bash
codegen generate
```

### For additional support and commands
```bash
codegen
```



## 📄 License

This project is licensed under the **MIT** License - see the [LICENSE](LICENSE) file for details.


## 👤 Author

**Kshitij Khandelwal**


## 🙏 Acknowledgments

- Thanks to all contributors who have helped shape this project
- Built with Go and Gin

---

<div align="center">
  
**[⬆ Back to Top](#codegen)**

*This README was automatically generated by CodeGen Pro with AI enhancement*

</div>
