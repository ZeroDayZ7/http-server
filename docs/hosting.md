pm2 restart kasandra-backend --update-env
history | less
history 20
pm2 logs kasandra-backend
pm2 start ./server --name "backend"

curl -v https://api.kaaa.ct8.pl/health -H "Origin: https://zerodayz7.github.io" --ssl-no-revoke
