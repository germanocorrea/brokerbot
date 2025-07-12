# BrokerBot

BrokerBot is a simple Telegram bot written in Go for sending IPC messages received through a UNIX socket to a Telegram chat.

> **Important!** This is a personal project , and a work in progress. Features may not be following standard conventions yet, and future updates may introduce breaking changes.

## Motivation and Target Users
This is designed as a simple solution for different programs and tasks running on personal computers or home servers to deliver notifications and messages in a unified and standalone way.
It is intended to be a lightweight solution for personal use cases with minimal overhead.

### Usage Examples:
- A program that tests internet connectivity and notifies you when the connection is restored
- A crawler searching for specific keywords on websites, RSS feeds, or other sources that notifies you when it finds matches
- Notifications when long-running tasks or processes complete
- System health monitoring alerts (disk space, temperature, etc.)

This is **not** meant to be a production-ready solution, nor should it be used in critical environments, alerting systems for production servers, or similar scenarios. Use in these environments is at your own risk.

## Prerequisites
- Go 1.24 or later
- A Telegram bot token (create one via [@BotFather](https://t.me/BotFather))
- ngrok for receiving messages from the bot 
- _Optional_: socat (if you want to send messages to the socket from the command line)

## Installing and Running
Installing isn't still in an idiomatic way for go programs, and this will be revised soon.
1. **Clone this repository:**
   ```bash
   git clone https://github.com/germanocorrea/brokerbot.git
   cd brokerbot
   ```

2. **Configure the application:**
   - The application uses command-line flags for configuration:
     ```
     -token        Telegram bot token (required)
     -password     Password to interact with the bot
     -socketPath   Path to the unix socket to listen to (defaults to $XDG_RUNTIME_DIR/brokerbot.sock or /tmp/brokerbot.sock)
     -ngrok        Use ngrok for the webhook (requires NGROK_AUTHTOKEN environment variable)
     -address      Webhook address to listen to (defaults to :8080)
     ```

3. **Start the server:**
   - You can build, or simply:
   ```bash
   go run main.go -token=foobar -ngrok
   ```

4. **Configure the webhook:**
   - If you're using ngrok (with the `--ngrok` flag), make sure the `NGROK_AUTHTOKEN` environment variable is set
   - The webhook will be automatically configured when the application starts
   - If not using ngrok, ensure your server is accessible from the internet and specify the address with `--address`

5. **Initialize chat reception:**
   - If you set a password with the `--password` flag, send a message with this password to your bot on Telegram to authenticate
   - Note: Since persistence isn't implemented yet, you'll need to do this every time the server restarts

6. **Send messages to the bot:**
   ```bash
   echo 'Hello from my system!' | socat - UNIX-CONNECT:"$XDG_RUNTIME_DIR/brokerbot.sock"
   ```

   If the `$XDG_RUNTIME_DIR` environment variable isn't set, the socket will be created at `/tmp/brokerbot.sock`

## Integration with Other Services

You can integrate BrokerBot with shell scripts, cron jobs, or any program capable of writing to a UNIX socket:

```bash
# Example shell script that notifies when a backup completes
backup_files /path/to/backup
echo "Backup completed successfully at $(date)" | socat - UNIX-CONNECT:"$XDG_RUNTIME_DIR/brokerbot.sock"
```

## Roadmap
- [x] HTTP server for Telegram bot webhook
- [x] Authentication system
- [x] Pooling for messages sent through UNIX socket
- [x] Improve configuration: replace the .env file with command-line arguments
- [x] Allow configuration of the server port
- [x] Integrate ngrok initialization directly into the application
- [ ] Allow choosing between webhook and message polling
- [ ] Persist chat IDs and other configurations (SQLite?)
- [ ] Add message formatting options
- [ ] Add commands for bot management

## Contributing
Contributions are welcome! Feel free to submit a Pull Request.
