# pubSubServer
Publish/Subscribe HTTP API Server, writen in GO and using Redis

## Usage

There are 4 main things you can do with this application:

- Publish a post to a topic
- Subscribe to a topic
- Unsubscribe to a topic
- Retrieve a post from a subscribed topic

When a user subscribes to a topic, they can retrieve all messages published to that topic from the point of subscription, not any from before that point. When they unsubscribe that place in the topics message que is lost, so re-subscription will start again from the most recent post. Any posts that have been read by all subscribed readers from the point at which it was posted will be deleted.


## API

### Subscribe: `POST` /[topic]/[user]

Adds a subscription to [topic] for [user], so the user will be able to read all posts to the topic that are published from that point onwards.

Response codes:
- 200: Subscription succeeded.

### Unsubscribe: `DELETE` /[topic]/[user]

Deletes a subscription to [topic] for [user], the user will not be able to read any posts from the topic without resubscribing.

Response codes:
- 200: Unsubscribe succeeded.
- 404: The subscription does not exist.

### Publish: `POST` /[topic]

Adds the HTTP request body to [topic] as a new post.

Request body: The message being published.

Response codes:
- 200: Publish succeeded.

### Retrieve: `GET` /[topic]/[user]

Gets the next post from [topic] for the [user], depending on what post they are currently on.

Response codes:
- 200: Retrieval succeeded.
- 204: There are no messages available for this topic on this user.
- 404: The subscription does not exist.

Response body: The body of the next message, if one exists.

## Memory Structure

As it may not be imediatly apparent in the code, I'll describe the data structure, I believe it to be about as simple and efficient as I could easily make it. As I'm using Redis thats important as its a "in memory" key/value store not to be a memory hog. There are 3 data structures produced by the server; User, Topic and Post:

### User `Hash`
- Key => User's Name (string), e.g. "Charlie"
  - Array of Field/Values => Topic (string) / Post Number (int), e.g. "Cats" / 5

### Topic `Key/Value`
- Key/Value => Topic (string) / Current Post (int), e.g. "Cats" / 7

### Post `Hash`
- Key => Topic + Post Number, e.g. "Cats6"
  -  Field/Value => User Count (string) / Number (int), e.g. "usrCount" / 5
  -  Field/Value => User Count (string) / Number (int), e.g. "content" / "I love cats: www.cats.com"

### future addition, "given more time"

- shard redis and add read only instances
- persist posts in sql DB or data store
- have more robust user system, perhaps oauth so others can't see you subs and read your posts
- post for subscription holders only, so no sub topics don't get messages auto deleted
