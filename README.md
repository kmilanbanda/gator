# Gator

Gator is an RSS feed aggregator programmed in Go. It will: 
 - allow users to Add RSS feeds from the internet 
 - store posts in a PostgreSQL database
 - follow/unfollow RSS feeds that other users added
 - view summariess of the aggregated posts in the terminal with a link

## Requirements
 - Go
 - Postgres

Ensure you have the latest [Go toolchain](https://golang.org/dl/) and a local Postgres database. From there, install using:

```bash
go install ...
```

## Configuration

Gator requires a config file in your home directory named `.gatorconfig.json`. Use the following structure for the file:

```json
{
    "db_url": "postgres://userrname:@localhost:5432/database?sslmode=disable"
}
```

You must replace the value with your database connection string. 

## Getting Started

After verifying you have the requirements and have configured your `.gatorconfig.json` file, you can begin with creating a user:

```bash
gator register <username>
```

Then you can add feeds:

```bash
gator addfeed <title> <url>
```

Start the aggregator to get posts:

```bash
gator agg 30s
```

View the posts:

```bash
gator browse [limit]
```

## Command List
 - `gator agg` (aggregates rss feeds)
 - `gator browse [limit]`
 - `gator register <username>` (registers user. Ex: register user1)
 - `gator login <username>` (logs into a user)
 - `gator follow <url>` (follows a feed)
 - `gator unfollow <url>` (unfollows a feed)
 - `gator users` (List all users)
 - `gator feeds` (List all feeds)
