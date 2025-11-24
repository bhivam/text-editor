## Basic Functionality Left
- Vertical Scrolling
- Line Wrap and/or Horizontal Scrolling
- Undo/Redo
- Modifiers
- Copy/Cut/Paste

## TCP Server
It is certainly true that you will need one routine for each connection. This routine will block on the TCP JSON decode and wait for data to stream in. I also have a second routine for every connection that reads in editor state data and then publishes it to the it's corresponding client.

There are two main fixes that I need:
1. There is no authority on what edits happen in what order or for what happens in a conflict situation. If two clients send the server edits while their editors are in the same state, how do we resolve that. We also don't have any resolution for how insertions and deletions affect the cursors of other connections. 
  - We need to have n + 1 routines where the extra routine only processes events sent to it.
  - Events need to come with an editor hash so we can quickly determine what state the editor is in and see if there is a conflict between two events that needs to be merged somehow.
  - Events need to come in with a time so that we can push them into a minheap and have the extra routine process them according to that timing.

There are other things I'm not considering like assigning priorities to event types. If a client is quiting, we can process them quiting first and any of their other events after, for example. Resize events can always be clubbed into the next editor update and can certainly be dropped after the individual editor is updated if there are other events needing to be processed. Et cetera.
