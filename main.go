package main

type model struct {
	choices  []string         // items on the todo list
	cursor   int              // which todo list item our cursor is pointing at
	selected map[int]struct{} // Which todo items are selected.
}
