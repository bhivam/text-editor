## My Text Editor

I'm making a text editor in Go using the gdamore/tcell package. I'm writing the frontend and backend of the application to be as separated as possible. Once the backend is sufficiently developed I would like to try rendering my own text using a graphics frame work. The goal is to get a vim-like editor. It will not be as complex or feature rich with additions like vimscript (at least thats not my goal right now).

Right now I'm using a basic Piece Table. This is just my starting data structure. I am assuming that this will mutate or even change entirely. This is because I can already see problems with representing logical lines and finding particular characters.
