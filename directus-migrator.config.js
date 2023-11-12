const config = {
  environments: [
    {
      name: "go-htmx-directus",
      endpoint: "https://go-htmx-directus.cookieserver.gg",
      accessToken: "***",
    },
    {
      name: "cookie-cloud",
      endpoint: "https://data.cookieserver.gg",
      accessToken: "***",
      production: true,
    },
  ],
};
module.exports = config;
