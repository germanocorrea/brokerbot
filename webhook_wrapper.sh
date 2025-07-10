echo -e "Finishing existing ngrok instances..."
pkill ngrok
source .env

echo -e "Initializing ngrok on 8080..."
ngrok http --log=stdout 8080 > /dev/null &
sleep 5
WEBHOOK_URL="$(curl http://localhost:4040/api/tunnels | jq ".tunnels[0].public_url")"
echo -e "WEBHOOK_URL=$WEBHOOK_URL"
echo -e "Sending new webhook to telegram bot..."
curl -F "url=$WEBHOOK_URL" https://api.telegram.org/bot"$TOKEN"/setWebhook
echo -e "\nAll done!"
#echo -e "\nPress CTRL+C to terminate..."
#( trap exit SIGINT ; read -r -d '' _ </dev/tty )