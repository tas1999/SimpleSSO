CREATE Table refresh_tokens (
  id SERIAL,
  user_id INTEGER REFERENCES users (Id),
  token VARCHAR(400),
  expiration timestamp with time zone
  );