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

TODO

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
