package main

import _ "github.com/lib/pq"

import (
	"github.com/kmilanbanda/gator/internal/config"
	"github.com/kmilanbanda/gator/internal/database"
	"github.com/google/uuid"
	"fmt"
	"log"
	"os"
	"database/sql"
	"context"
	"time"
	"net/http"
	"html"
	"io"
	"encoding/xml"
	"strconv"
)

type state struct {
	db	*database.Queries
	cfgPtr	*config.Config
}

type command struct {
	name	string
	args	[]string
}

type commands struct {
	commandMap	map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	err := c.commandMap[cmd.name](s, cmd)
	if err != nil {
		return fmt.Errorf("Error occured when running command: %w", err)
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f	
}

//for ensuring a user is logged in
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {	
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfgPtr.CurrentUser)
		if err != nil {
			return fmt.Errorf("Error getting user: %w", err)
		}

		err = handler(s, cmd, user)
		if err != nil {
			return fmt.Errorf("Error in handler: %w", err)
		}

		return nil
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Error: the login command expects a single argument.")
	}

	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == sql.ErrNoRows {
		return fmt.Errorf("User does not exist.")
	} else if err != nil {
		return fmt.Errorf("User not found.")
	}

	err = s.cfgPtr.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error during handler login: %w", err)
	}
	fmt.Printf("User has been set to %s\n", cmd.args[0])

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Error: the register command expects a single argument.")
	}
	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == nil {
		return fmt.Errorf("Error: user already exists")
	} else if err == sql.ErrNoRows {
		
	} else {
		return fmt.Errorf("Error: %w", err)
	}

	params := database.CreateUserParams{
		ID:		uuid.New(),
		CreatedAt:	time.Now(),
		UpdatedAt:	time.Now(),
		Name:		cmd.args[0],
	}

	user, err = s.db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("Error during user creation: %w", err)
	}

	err = s.cfgPtr.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error setting user: %w", err)
	}
	fmt.Printf("Created user: %+v\n", user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.args) > 0 {
		return fmt.Errorf("Error: the reset command expects no arguments.")
	}

	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error resetting users: %w", err)
	}

	return nil
}

func handlerUsers(s *state, cmd command) error {
	
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting users: %w", err)
	}

	for _, user := range users {
		fmt.Printf(" * %v", user)
		if s.cfgPtr.CurrentUser == user {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Error: expecting 1 argument for agg command")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error setting request delay: %w", err)
	}
	fmt.Printf("Collecting feed every %s\n", cmd.args[0])

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			log.Printf("Error scraping feeds: %w", err)
		}
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("Error: expecting 2 arguments for addfeed command (name, url)")
	}

	feedParams := database.CreateFeedParams{
		ID:		uuid.New(),
		CreatedAt:	time.Now(),
		UpdatedAt:	time.Now(),
		Name:		cmd.args[0],
		Url:		cmd.args[1],
		UserID:		user.ID,
	}
	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return fmt.Errorf("Error creating feed object: %w", err)
	}

	fmt.Printf("Created feed for %s: %+v\n", user.Name, feed)

	feedFollowParams := database.CreateFeedFollowParams{
		ID:		uuid.New(),
		CreatedAt:	time.Now(),
		UpdatedAt: 	time.Now(),
		UserID:		user.ID,
		FeedID:		feed.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return fmt.Errorf("Error creating feed-follow record: %w", err)
	}
	
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf(" * \"%s\": %s (%s)\n", feed.Name, feed.Url, feed.Name_2)
		fmt.Printf("\n")
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Error: follow command expects url")
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error getting feed by url: %w", err)
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:		uuid.New(),
		CreatedAt:	time.Now(),
		UpdatedAt: 	time.Now(),
		UserID:		user.ID,
		FeedID:		feed.ID,	
	}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return fmt.Errorf("Error creating feed-follow record: %w", err)
	}

	fmt.Printf("New Feed: \"%s\" (%s)", feedFollow.FeedName, feedFollow.UserName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
	return fmt.Errorf("Error getting follows: %w", err)
	}

	for _, follow := range follows {
	fmt.Printf(" * %s", follow.FeedName)
	}

	return nil	
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Error: Unfollow expects a url.")
	}

	deleteParams := database.DeleteFeedFollowParams{
		Url:	cmd.args[0],
		UserID: user.ID,
	}
	err := s.db.DeleteFeedFollow(context.Background(), deleteParams)
	if err != nil {
		return fmt.Errorf("Error unfollowing feed: %w", err)
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	var err error
	if len(cmd.args) >= 1 {
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("Error during string conversion Atoi(): %w", err)
		}
	}

	getPostsForUserParams := database.GetPostsForUserParams{
		UserID:		user.ID,
		Limit:		int32(limit),
	}
	posts, err := s.db.GetPostsForUser(context.Background(), getPostsForUserParams)
	if err != nil {
		return fmt.Errorf("Error getting posts for user: %w", err)
	}

	for _, post := range posts {
		fmt.Printf(" * %s (%s) \n     - %s\n", post.Title, post.Url, post.Description)
	}

	return nil
}

type RSSItem struct {
	Title		string `xml:"title"`
	Link		string `xml:"link"`
	Description 	string `xml:"description"`
	PubDate		string `xml:"pubDate"`
}

type RSSFeed struct {
	Channel struct {
		Title		string		`xml:"title"`
		Link		string		`xml:"link"`
		Description	string		`xml:"description"`
		Item		[]RSSItem	`xml:"item"`
	} `xml:"channel"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating new request: %w", err)
	}

	req.Header.Set("User-Agent", "gator")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error making new request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading request body: %w", err)
	}

	var feed RSSFeed
	if err = xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("Error unmarshaling xml: %w", err)
	}

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	return &feed, nil
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting next feed to fetch: %w", err)
	}
	
	fetchTime := sql.NullTime{
		Time:	time.Now(),
		Valid:	true,
	}
	markFeedFetchedParams := database.MarkFeedFetchedParams{
		ID:		feed.ID,
		LastFetchedAt:	fetchTime,
	}
	err = s.db.MarkFeedFetched(context.Background(), markFeedFetchedParams)
	if err != nil {
		return fmt.Errorf("Error updating feed last_fetched_at: %w", err)
	}

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("Error fetching feed: %w", err)
	}

	for _, item := range rssFeed.Channel.Item {
		description := sql.NullString{
			String:		item.Description,
			Valid:		true,
		}

		published, err := parseTime(item.PubDate) 
		if err != nil {
			log.Printf("Error parsing published date: %w", err)
		}

		createPostParams := database.CreatePostParams{
			ID:		uuid.New(),
			CreatedAt:	time.Now(),
			UpdatedAt:	time.Now(),
			Title:		item.Title,
			Url:		item.Link,
			Description:	description,
			PublishedAt:	published,
			FeedID:		feed.ID,
		}
		_, err = s.db.CreatePost(context.Background(), createPostParams)
		if err != nil {
			log.Printf("Error: %w", err)
		}
	}

	return nil
}

func parseTime(timeStr string) (time.Time, error) {
    layouts := []string{
        time.RFC3339,
        time.RFC822,
        time.RFC1123,
        "2006-01-02 15:04:05",
        "02 Jan 2006 15:04:05 MST",
    }
    
    for _, layout := range layouts {
        if t, err := time.Parse(layout, timeStr); err == nil {
            return t, nil
        }
    }
    
    return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading in main(): %v", err)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	dbQueries := database.New(db)

	var s state
	s.db = dbQueries
	s.cfgPtr = &cfg
	
	var cmds commands
	cmds.commandMap = make(map[string]func(*state, command) error)

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	if len(os.Args) < 2 {
		log.Fatalf("Error: no arguments")
	}
	cmd := command{
		name: 	os.Args[1],
		args: 	os.Args[2:],
	}

	err = cmds.run(&s, cmd)
	if err != nil {
		log.Fatalf("Error running command:\n--- %v", err)
	}

	//fmt.Printf("Config db_url: %s\nConfig current_user_name: %s\n", cfg.DbUrl, cfg.CurrentUser)
}
