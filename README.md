# PHP Version Manager (phpvm)

> **⚠️ Aviso / Notice**: Este es un proyecto personal en desarrollo activo. No está destinado para uso en producción. El código fuente tiene restricciones de uso - consulta la [LICENCIA](LICENSE) para más detalles. / This is a personal project under active development. Not intended for production use. Source code has usage restrictions - see [LICENSE](LICENSE) for details.

A command-line tool to manage multiple PHP versions on your system.

## Features

- List available PHP versions
- Install specific PHP versions
- Switch between PHP versions
- Check current PHP version

## Next Features

- Create binaries for PHP versions
- Create command for install PHP extensions

## Installation

1. Make sure you have Go installed (Go 1.16+ required)
2. Clone this repository
3. Build the project:
   ```bash
   go build -o phpvm
   ```
4. Move the binary to your PATH:
   ```bash
   sudo mv phpvm /usr/local/bin/
   ```

## Usage

### List available PHP versions
```bash
phpvm list
```

### Install a PHP version
```bash
phpvm install 8.2.0
```

### Show current PHP version
```bash
phpvm version
```

### Set PHP version
```bash
phpvm version 8.2.0
```

## Requirements

- Linux/macOS (Windows support coming soon)
- Build tools (gcc, make, autoconf, libtool, automake)
- Git
- wget or curl

## License

This project is licensed under a proprietary license with usage restrictions. See the [LICENSE](LICENSE) file for complete terms and conditions.

**Important**: The source code is provided for educational purposes only. Modification, distribution, and commercial use of the source code are prohibited.
