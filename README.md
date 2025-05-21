# Workbench

Welcome to Workbench! This is my personal collection of command-line tools for automating everyday tasks and storing useful utilities I don't want to forget. It's also a personal project to expand my Go programming knowledge and skills.

## Installation

```sh
$ go install github.com/bdunn313/workbench@latest
```

## Usage

### Roll

Roll is a simple dice roller. It takes a dice notation and rolls the dice. It supports the following notation:

- `d6`: Roll a single 6-sided die
- `2d6`: Roll two 6-sided dice
- `d6+1`: Roll a single 6-sided die and add 1 to the result
- `2d6+1`: Roll two 6-sided dice and add 1 to the result
- `2d4+3d8+4+1d20+2`: Roll two 4-sided dice, three 8-sided dice, a single 20-sided die, and add 4 and 2 to the result

```sh
$ workbench roll 2d4+3d8+4+1d20+2
Rolling: 2d4+3d8+4+1d20+2

 2      2d4
30    3d8+4
15   1d20+2
------------
47
```

It also supports plain formatting so you can pipe the output

```sh
$ workbench roll 2d4+3d8+4+1d20+2 --plain | cat
47
```

**Note:** Right now it does not support negative numbers for modifiers. That is tracked [here](https://github.com/bdunn313/workbench/issues/1).

### Table

Table is a command for randomly selecting rows from CSV files. It supports the following features:

- Read from a file or stdin
- Skip header row automatically
- Output in either formatted or plain CSV format

```sh
# Roll on a table from a file
$ workbench table roll path/to/table.csv

# Roll on a table from stdin
$ cat table.csv | workbench table roll -

# Get plain CSV output
$ workbench table roll path/to/table.csv --plain
```

The formatted output will show each column with its header (if present) or column number, while the plain output will be comma-separated values suitable for piping to other commands.

### Prepare

Prepare helps you get ready for your upcoming week by:

- Pulling calendar events and tasks from your Google account
- Asking a series of questions about your upcoming week
- Analyzing your commitments and providing a summary

```sh
$ workbench prepare
```

Before using the prepare command, you'll need to set up Google OAuth credentials:

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Calendar and Tasks APIs:
   - Go to "APIs & Services" > "Library"
   - Search for "Google Calendar API" and enable it
   - Search for "Google Tasks API" and enable it
4. Create OAuth 2.0 credentials:
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth client ID"
   - Choose "Desktop app" as the application type
   - Give it a name and click "Create"
   - You'll get a client ID and client secret

5. Configure your credentials in [`~/.workbench.yaml`](.workbench.example.yaml):

```yaml
google:
  client_id: "your-client-id.apps.googleusercontent.com"
  client_secret: "your-client-secret"
  token_file: "~/.workbench/google_token.json"
```

The first time you run the command, it will:
1. Open a browser window for you to authorize the application
2. Ask you to paste the authorization code
3. Store the token for future use

The command will then:
1. Show your upcoming calendar events for the next week
2. Ask you a series of questions about your upcoming week
3. Analyze your commitments and provide a summary
