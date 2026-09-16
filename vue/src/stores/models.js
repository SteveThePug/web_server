/**
 * Hand-written classes that coerce raw API JSON into real Date objects.
 *
 * STATUS: not imported anywhere. Stores consume the GraphQL payloads directly and
 * format dates at the point of display, so these are dormant; kept as a record of
 * the intended shape of User/Message/Post.
 */

export class User {
  constructor({ id, createdAt, updatedAt, deletedAt, username, admin }) {
    this.id = id;
    this.createdAt = new Date(createdAt);
    this.updatedAt = new Date(updatedAt);
    this.deletedAt = deletedAt ? new Date(deletedAt) : null;
    this.username = username;
    this.admin = admin;
  }
}

export class Message {
  constructor({ id, text, author, createdAt, deletedAt }) {
    this.id = id;
    this.content = text;
    this.author = author ? new User(author) : null;
    this.createdAt = new Date(createdAt);
    this.deletedAt = deletedAt ? new Date(deletedAt) : null;
  }
}

export class Post {
  constructor({
    id,
    title,
    author,
    authorID,
    content,
    createdAt,
    updatedAt,
    deletedAt,
  }) {
    this.id = id;
    this.title = title;
    this.authorID = authorID;
    this.author = author ? new User(author) : null;
    this.content = content;
    this.createdAt = new Date(createdAt);
    this.updatedAt = new Date(updatedAt);
    this.deletedAt = deletedAt ? new Date(deletedAt) : null;
  }
}

// Utility function to parse posts from API
export function parsePosts(postsArray) {
  if (!Array.isArray(postsArray)) return [];
  return postsArray.map((post) => new Post(post));
}
