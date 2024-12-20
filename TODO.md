## Basic Functionality Left
- Vertical Scrolling
- Line Wrap and/or Horizontal Scrolling
- Undo/Redo
- Modifiers

## TCP Remote Editing 
The goal is to have a server that can be accessed via SSH by multiple users simultaneously. The server will have one file open that users can edit as they please. We need some basics:
- Some way to specify either key strokes or commands to the server. I'm thinking it should just be key strokes.
- The server should return back the computed state of the file after each key stroke.
- Once this is done for one user, we need to be able to take in streams of key strokes from multiple users at once and reconcile them. Whenever one user makes a change, we want that change to be propogated to everyone. Our first, simple strategy will be to just take in the key strokes and apply them in the order they are recieved. We will transmit the new state upon each key stroke.

## Client Design
There are two modes of operation: local and remote. In the local mode, we are self hosting the backend on the client side and editing a file locally.

In the remote end, the client is connecting to the server.

We need to identify what information the client requires in order to render. There is a question of whether rendering should be server side. Would it be more expensive to serialize entire state and send that to the client or have the server rendering and sending. I will do the state serialization for now.
