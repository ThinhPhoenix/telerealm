# telerealm 📡

A lightweight Content Delivery Network (CDN) that leverages the Telegram Bot API for file storage and retrieval.

## Overview

telerealm provides a cost-effective solution for hosting files and serving them via a CDN-like infrastructure using Telegram's robust backend. This project is perfect for those looking for an affordable way to store and retrieve files without managing complex infrastructure.

![image](https://github.com/user-attachments/assets/eb82385e-5eeb-4770-8951-2424841db4f4)

**Author:** Lai Chi Thinh (ThinhPhoenix) - FPT University

## Features

- **File Upload**: Upload files to Telegram Bot API and receive a unique file URL for accessing the content
- **File Metadata Retrieval**: Retrieve file URLs and metadata (such as file size) using dedicated API endpoints
- **CORS Support**: Seamless integration with web applications through Cross-Origin Resource Sharing configuration
- **Secure File Download**: Generate secure, unique URLs for file downloads with automatic expiration
- **Bot and Chat Verification**: Verify bot and chat information to ensure proper configuration

## Advantages

### Pros ✅

- **Unlimited Storage**: Utilizes Telegram's cloud storage for files, offering virtually unlimited capacity
- **Easy to Use**: Integration with Telegram Bot API simplifies file upload and retrieval operations
- **Free**: No additional cost for storage or bandwidth usage beyond what Telegram charges for bot API usage

### Cons ⚠️

- **Automatic Removal**: Files may be removed if Telegram considers them inactive for a prolonged period due to privacy policies and storage management

## Getting Started

### Prerequisites

- Go programming language (version 1.16 or later)
- Telegram Bot API token (obtain one by creating a new bot using BotFather)

### Installation

1. **Create a bot in Telegram**:

   - Follow the instructions on the [Telegram Bot Features](https://core.telegram.org/bots/features) page to create a new bot and obtain your API token

2. **Clone the repository**:

   ```bash
   git clone https://github.com/ThinhPhoenix/telerealm.git
   cd telerealm
   ```

3. **Installed dependencies**:

   ```bash
   go get github.com/gin-contrib/cors
   go get github.com/gin-gonic/gin
   go get github.com/google/uuid
   go get github.com/joho/godotenv
   ```

4. **Run the project**:
   ```bash
   go run main.go
   ```

The server will start running on http://localhost:7777 by default.

## Usage

### Upload a File

To upload a file, send a POST request to the `/send` endpoint with the following form data:

- `chat_id`: The chat ID where you want to upload the file (you can use your own chat ID or a group/channel ID)
- `document`: The file you want to upload

**Example**:

```bash
curl -X POST -H "Authorization: Bearer <your_bot_token>" -F "chat_id=<your_chat_id>" -F "document=@/path/to/your/file" http://localhost:7777/send
```

### Get File URL

To retrieve the file URL, send a GET request to the `/url` endpoint with the following query parameters:

- `file_id`: The file ID obtained from the upload response

**Example**:

```bash
curl -X GET -H "Authorization: Bearer <your_bot_token>" "http://localhost:7777/url?file_id=<your_file_id>"
```

### Retrieve File

You can download the file by accessing the secure URL generated after uploading:

- `/drive/:id`: Endpoint to download the file associated with `:id` (secure ID)

**Example**:

```bash
curl -OJL http://localhost:7777/drive/<secure_id>
```

### Get File Information

To get information about a file (including its size and URL), send a GET request to the `/info` endpoint with the following query parameters:

- `file_id`: The file ID obtained from the upload response

**Example**:

```bash
curl -X GET -H "Authorization: Bearer <your_bot_token>" "http://localhost:7777/info?file_id=<your_file_id>"
```

### Verify Bot and Chat

To verify bot and chat information, send a GET request to the `/verify` endpoint with the following query parameters:

- `chat_id`: The chat ID where you want to verify the bot

**Example**:

```bash
curl -X GET -H "Authorization: Bearer <your_bot_token>" "http://localhost:7777/verify?chat_id=<your_chat_id>"
```

## Important Notes

- Replace `<your_bot_token>`, `<your_chat_id>`, `<your_file_id>`, and `<secure_id>` with actual values as per your application's requirements
- Ensure that the Telegram Bot API token is kept secure and not exposed in client-side code
- Files may be subject to Telegram's inactive file removal policies

## Security Considerations

- Keep your bot token private and secure
- Consider implementing rate limiting for your API endpoints
- Monitor your usage to ensure compliance with Telegram's terms of service

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
