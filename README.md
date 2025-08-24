# GoLoad

GoLoad is a flexible and powerful HTTP load testing library for Go, designed to help you perform stress testing and benchmarking of your web applications.

## Features

- Support for multiple HTTP methods (GET, POST, PUT, DELETE, HEAD, PATCH)
- Configurable virtual users (VU) and execution timepoints
- Custom user agent simulation (Chrome, Firefox, Safari, Edge, Opera, IE, Android, iOS)
- Request/Response handling with headers, cookies, and body support
- Configurable test duration and timeout settings
- Optional logging with custom output path
- Flexible execution mode configuration


### Available HTTP Methods

- `GET`
- `POST`
- `PUT`
- `DELETE`
- `HEAD`
- `PATCH`

### Supported User Agents

- `ChromeAgent`
- `FirefoxAgent`
- `SafariAgent`
- `EdgeAgent`
- `OperaAgent`
- `IEAgent`
- `AndroidAgent`
- `IOSAgent`

### Execution Timepoints

Use execution timepoints to configure how the number of virtual users changes over time:
