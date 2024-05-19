package main

type Editor struct {
  content Content
  cursor Cursor
}

func initialize_editor(path string) Editor {
  cursor := Cursor{index: 0}
  content := Content{}

  content.load_from_file(path)
  
  editor := Editor{content: content, cursor: cursor}
  
  return editor 
}
