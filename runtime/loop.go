package runtime

// Loop represents the $loop variable available in foreach/for loops
type Loop struct {
	Index     int   // Current iteration index (0-based)
	Iteration int   // Current iteration number (1-based)
	Remaining int   // Remaining iterations
	Count     int   // Total count of items (-1 if unknown)
	First     bool  // Is this the first iteration?
	Last      bool  // Is this the last iteration?
	Even      bool  // Is this an even iteration?
	Odd       bool  // Is this an odd iteration?
	Depth     int   // Loop nesting depth (1-based)
	Parent    *Loop // Parent loop (for nested loops)
}

// LoopStack manages nested loop contexts
type LoopStack struct {
	stack []*Loop
}

// NewLoopStack creates a new loop stack
func NewLoopStack() *LoopStack {
	return &LoopStack{
		stack: make([]*Loop, 0),
	}
}

// NewLoop creates a new Loop instance
// count: total number of items (-1 if unknown, e.g., for while loops)
// depth: the nesting depth
func NewLoop(count, depth int) *Loop {
	return &Loop{
		Index:     -1,
		Iteration: 0,
		Remaining: count,
		Count:     count,
		First:     true,
		Last:      count == 1,
		Even:      false,
		Odd:       true,
		Depth:     depth,
		Parent:    nil,
	}
}

// Update updates the loop for the next iteration
func (l *Loop) Update(index int) *Loop {
	newLoop := &Loop{
		Index:     index,
		Iteration: index + 1,
		Count:     l.Count,
		Depth:     l.Depth,
		Parent:    l.Parent,
	}

	if l.Count >= 0 {
		newLoop.Remaining = l.Count - index - 1
		newLoop.Last = index == l.Count-1
	} else {
		newLoop.Remaining = -1
		newLoop.Last = false
	}

	newLoop.First = index == 0
	newLoop.Even = (index+1)%2 == 0
	newLoop.Odd = (index+1)%2 == 1

	return newLoop
}

// Push pushes a new loop onto the stack
func (s *LoopStack) Push(loop *Loop) {
	if len(s.stack) > 0 {
		loop.Parent = s.stack[len(s.stack)-1]
	}
	loop.Depth = len(s.stack) + 1
	s.stack = append(s.stack, loop)
}

// Pop removes and returns the top loop from the stack
func (s *LoopStack) Pop() *Loop {
	if len(s.stack) == 0 {
		return nil
	}
	loop := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return loop
}

// Current returns the current (top) loop without removing it
func (s *LoopStack) Current() *Loop {
	if len(s.stack) == 0 {
		return nil
	}
	return s.stack[len(s.stack)-1]
}

// Depth returns the current nesting depth
func (s *LoopStack) Depth() int {
	return len(s.stack)
}
