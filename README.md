# GoGoScrap

This bot is doing scrapping routine for you. It finds your website rank in Google Search. 

# Usage

Just send a message: `<query>, <website>`.

Bot also supports long searches by one particular website. If you need to do a multiple searches, use the following command: `/starturlsearch <website>`.

After that you will be able to just send a query, without mentioning your website everytime. Don't forget to stop this search by using `/stopurlsearch` command.

## Building

`docker build -t gogoscrap .`

## Running

`docker run gogoscrap -env BOT_TOKEN=<your bot token>`

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
