supernova-go-bot is a Telegram bot based on [Go Telegram API framework](https://github.com/go-telegram/bot). It integrates with various services like OpenAI and AccuWeather with authorization middleware so that only promoted users can have an access to the fetching commands. The database is managed by PostgreSQL and is used to store promoted users and to record the users' activity. The user can be promoted by the admin, which id among other settings like API keys for the services and db connection string is stored in the `.env` file.

## Installation

0. Install Go and set up your workspace. Install make if on Windows via chocolatey:
    ```sh
    choco install make
    ```

1. Create the bot on Telegram with BotFather and save its API key.

2. Clone this repository

3. Install dependencies:
    ```sh
    make dep
    ```

4. Create and fill the env file (as in [example file](.env.example)) with the required parameters. If you do not fill some external tokens, the commands that require their APIs won't be available.

5. Install PostgreSQL and create a database. Fill the env file with the database connection string parameters. 

6. Execute the up migration to create the tables in the database:
    ```sh
    make migrate-up
    ```

7. Run the bot:
    ```sh
    make run
    ```

8. Type /help in the bot chat to see the available commands

## Usage

The bot has the following commands:

### Regular commands
- /help - shows the available commands
- /getid - shows the user's Telegram ID, useful for the admin to promote users

### Promoted commands
- /weather <city> <N> days|hours - fetches the weather forecast for the city for the next N days or hours from AccuWeather
- /chat <prompt> - sends the prompt to OpenAI and returns the response

### Admin commands
- /allow <user_id> - promotes the user with the given ID to have access to the promoted commands

## License
The project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more information.

## Author
The project was created by Nikita Smirnov. You can contact me via [email](mailto:detectivecolombo@gmail.com) or by opening an issue in this repository.