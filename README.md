# Fondazione FS News Bot (Not ufficial)

This telegram bot keeps updated its users with the newest historic trains from Fondazione FS.
Trains are updated every hour and only train in the near(configurable) future will be sent.

### Config
Create a file called `config.json` (you can copy `config.json.example`).

`TelegramBotToken` is the token given to you by [@BotFather](t.me/BotFather), `ChannelId` is the userid of the channel this can be obtained from a message in the channel, `TrainsUntil{Year,Month,Day}sInTheFuture` are used to select the last day in the future in which scheduled train will be sent(a train coming after that time will not be sent yet), set at least one to a negative number to disable the feature, defaults to 1 month.