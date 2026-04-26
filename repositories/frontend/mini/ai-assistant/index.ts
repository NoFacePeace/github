import { routes } from "./routes/routes";
const server = Bun.serve({
  port: 80,
  routes: {
    "/": () => new Response('Bun!'),
    ...routes
  }
});

console.log(`Listening on ${server.url}`);