# Stop Screenshots, Start Sharing: How We Open-Sourced Our Solution to Claude's Visualisation Problem

Ever created a stunning data visualisation in Claude Desktop, only to realise you can't share it with anyone who doesn't have Claude?

We've all been there. You ask Claude to whip up a beautiful interactive chart. It delivers brilliantly—responsive graphs, perfect colour schemes, the works. Then your colleague asks, "Can you send that to me?"

Cue the awkward screenshot dance.

## The Problem Nobody Talks About

Here's what typically happens:

You spend 10 minutes getting Claude to create the perfect visualisation. Another 5 minutes tweaking it. Then someone asks to see it, and you're stuck taking screenshots like it's 2005.

It's not just inconvenient—it defeats the entire purpose of creating *interactive* visualisations.

That's why we built MCP HTML Share.

## Enter the Model Context Protocol (MCP)

Before we dive into our solution, let's talk about MCP—the unsung hero of the AI tool ecosystem.

MCP is Anthropic's protocol that lets AI assistants like Claude interact with external tools and services. Think of it as giving Claude superpowers beyond just generating text. With MCP, Claude can:

- Query databases
- Call APIs
- Create files
- And now, thanks to our tool, share HTML content via URLs

It's the bridge between AI's creative capabilities and the real world's practical needs.

## Our Solution: MCP HTML Share

We've open-sourced a dead-simple MCP server that does one thing brilliantly: takes HTML content from Claude and gives you back a shareable URL.

Here's how it works:

**Claude creates your visualisation** → **Our tool uploads it to Google Cloud Storage** → **You get a shareable URL**

No screenshots. No "Can you install Claude?" conversations. Just a link that works.

Need to keep things internal? No problem. The tool supports both public URLs (for sharing with the world) and signed URLs (for keeping visualisations within your organisation). You choose based on your security needs.

## Setting It Up Locally (It's Easier Than You Think)

Want to try it yourself? Here's the quickest path:

### 1. Using Docker (Recommended)

```bash
docker run -p 8080:8080 \
  -e GOOGLE_APPLICATION_CREDENTIALS=/creds/credentials.json \
  -v /path/to/credentials.json:/creds/credentials.json \
  ghcr.io/loveholidays/mcp-html-share:latest \
  --bucket=your-bucket-name
```

### 2. Configure Claude Desktop

Add this to your Claude Desktop config:

```json
{
  "mcpServers": {
    "html-share": {
      "command": "npx",
      "args": ["mcp-remote", "http://localhost:8080"]
    }
  }
}
```

That's it. Claude can now share visualisations with the world.

## Taking It to Production with Kubernetes

For teams wanting to deploy this at scale, Kubernetes deployment is straightforward. The service runs as a stateless deployment with horizontal scaling capabilities. Simply:

1. **Deploy the container** with your GCS credentials mounted as a secret
2. **Configure the bucket** and choose between public or authenticated URLs
3. **Expose the service** through your ingress controller
4. **Monitor health** via the built-in `/livez` endpoint and Prometheus metrics at `/metrics`

The beauty is in the simplicity—no complex state management, no database dependencies. Just a lightweight Go service that scales horizontally as your team grows.

Pro tip: Use the `--public-url=false` flag for internal-only sharing. Perfect for sensitive data visualisations that should stay within your organisation's Google Workspace.

## Why This Matters

This isn't just about convenience (though that's nice too).

It's about unlocking the full potential of AI-generated content. When visualisations are trapped in screenshots, they lose their interactivity, their responsiveness, and frankly, their magic.

By making sharing frictionless, we're removing one more barrier between AI capabilities and real-world impact.

## The Technical Bits (For the Curious)

Built in Go for blazing-fast performance. Uses Google Cloud Storage for reliability.

The entire codebase is open source (LGPL-3.0), well-tested, and production-ready. We've been dogfooding it internally for months.

## Join Us in Making AI More Shareable

Here's what you can do right now:

**Try it out**: Clone the repo and get it running in under 5 minutes  
**Star the project**: Help others discover it on [GitHub](https://github.com/loveholidays/mcp-html-share)  
**Contribute**: We welcome PRs, especially for supporting other cloud providers  
**Share your creations**: Tag us when you share your first visualisation

Because the best visualisations are the ones people actually see.

---

*Ready to liberate your visualisations? Head to [github.com/loveholidays/mcp-html-share](https://github.com/loveholidays/mcp-html-share) and give it a spin. Your colleagues will thank you.*