# Log Listener Program

## Overview

This log listener program is designed to listen to log data over UDP or TCP, process the logs, and store them in a Manticore database. The configuration is managed through a YAML file, and the application supports multiple log configurations. The main components of the program are the configuration loader, log listener, log processor, and log parser.

## Features

- Supports UDP and TCP protocols for log listening.
- Processes log entries using a configurable worker pool.
- Supports JSON and non-JSON log formats.
- Inserts log entries into Manticore database tables.
- Configurable through a YAML file.

## Table of Contents

- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Usage](#usage)
- [Main Components](#main-components)
  - [Configuration Loader](#configuration-loader)
  - [Log Listener](#log-listener)
  - [Log Processor](#log-processor)
  - [Log Parser](#log-parser)
- [Dependencies](#dependencies)
- [Contributing](#contributing)
- [License](#license)

## Installation

### Prerequisites

- Go 1.15 or higher
- Manticore Search
- GNU Make

### Installation

Clone the repository:

```sh
git clone <repository-url>
cd logListener
```

### Build the Application
```
make build
```


