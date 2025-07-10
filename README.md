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

1. **Clone this repository:**
   ```bash
   git clone https://github.com/germanocorrea/brokerbot.git
   cd brokerbot
   ```

2. **Configure the environment:**
   - Create an `.env` file using `example.env` as a template
   - Add your Telegram bot token and comma-separated list of allowed usernames:
     ```
     TOKEN=your_telegram_bot_token
     USERS_ALLOW_LIST=username1,username2
     ```

3. **Start the server:**
   - You can build, or simply:
   ```bash
   go run main.go
   ```

4. **Configure the webhook:**
   - Run the bash script to expose the local server and configure the Telegram webhook automatically:
     ```bash
     ./webhook_wrapper.sh
     ```

5. **Initialize chat reception:**
   - Send a message containing only "marco" to your bot on Telegram
   - When the bot replies with "Polo!!", your chat is registered to receive messages
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
- [x] Simple authentication through username allowlist
- [x] Pooling for messages sent through UNIX socket
- [ ] Allow choosing between webhook and message polling
- [ ] Improve configuration: replace the .env file with command-line arguments
- [ ] Allow configuration of the server port
- [ ] Persist chat IDs and other configurations (SQLite?)
- [ ] Integrate ngrok initialization directly into the application
- [ ] Add message formatting options
- [ ] Add commands for bot management

## Contributing
Contributions are welcome! Feel free to submit a Pull Request.