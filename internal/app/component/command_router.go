/*
COMMAND ROUTER

	Mixin that handles type assertions and global intercept logic
	for update commands.

	[UpdateChan] ──▶ [Router] ──┬──▶ [Global Handler] (Base)
								└──▶ [Typed Handler]  (Specific)
*/
package component

// CommandHandlerFunc is a function that handles a specific command type
type CommandHandlerFunc func(cmd UpdateCmd)

// TypedCommandHandlerFunc is a generic function that handles a command with typed data.
type TypedCommandHandlerFunc[T any] func(data T)

/*
API
┌────────────────────┬────────────────┬────────────────────────────┐
│ Requirement        │ Method         │ Behavior                   │
├────────────────────┼────────────────┼────────────────────────────┤
│ Standard Logic     │ OnCommand      │ Stops if Base handles it   │
│ Side Effects       │ OnCommandRaw   │ Runs even if Base handles  │
│ Typed Data (Safe)  │ OnCommandTyped | Auto-casts Data to T       │
│ Typed Side Effects │ OnCmdTypedRaw  | Typed + Runs always        │
└────────────────────┴────────────────┴────────────────────────────┘
*/
type CommandRouter struct {
	component interface {
		RegisterCommandHandler(cmdType any, handler func(UpdateCmd))
		HandleGlobalUpdate(cmd UpdateCmd) bool
	}
}

// Init initializes the CommandRouter with a reference to the component.
func (cr *CommandRouter) Init(comp interface {
	RegisterCommandHandler(cmdType any, handler func(UpdateCmd))
	HandleGlobalUpdate(cmd UpdateCmd) bool
}) {
	cr.component = comp
}

/*
ROUTE LOGIC
      ┌──────────────────┐        ┌──────────────────┐
      │    OnCommand     │        │   OnCommandRaw   │
      └────────┬─────────┘        └────────┬─────────┘
               ▼                           ▼
      ┌──────────────────┐        ┌──────────────────┐
      │  Global Handle?  │        │  Global Handle?  │
      └────┬────────┬────┘        └────────┬─────────┘
       Yes │        │ No                   │ (Run Both)
           ▼        ▼                      ▼
        [STOP]   [RUN HANDLER]        [RUN HANDLER]
     (Handled)   (Custom Logic)       (Side Effects)
*/

// OnCommand registers a handler function for a specific command type.
func (cr *CommandRouter) OnCommand(cmdType any, handler CommandHandlerFunc) {
	if cr.component == nil {
		panic("CommandRouter not initialized - call Init() first")
	}

	// Wrap the handler to check for global commands first
	wrappedHandler := func(cmd UpdateCmd) {
		// Try to handle as global command first
		if cr.component.HandleGlobalUpdate(cmd) {
			return
		}
		// Otherwise, call the custom handler
		handler(cmd)
	}

	cr.component.RegisterCommandHandler(cmdType, wrappedHandler)
}

// OnCommandRaw registers a handler that will be called for ALL commands of this type,
// even if they're global commands.
func (cr *CommandRouter) OnCommandRaw(cmdType any, handler CommandHandlerFunc) {
	if cr.component == nil {
		panic("CommandRouter not initialized - call Init() first")
	}

	// Wrap handler to always try global commands first, then call custom handler
	wrappedHandler := func(cmd UpdateCmd) {
		cr.component.HandleGlobalUpdate(cmd)
		handler(cmd)
	}

	cr.component.RegisterCommandHandler(cmdType, wrappedHandler)
}

/*
GENERIC UNBOXING
Automatically cast interface{} data to concrete types.

	Input: UpdateCmd 				  Handler: func(data T)
 	┌───────────────────┐             ┌───────────────────┐
 	│ Cmd: CmdSetColor  │             │ T: imgui.Vec4     │
 	│ Data: interface{} │────[ T ]───▶│ Data: data.X, (R) │
 	└───────────────────┘             │       data.Y, (G) │
                                      │       data.Z, (B) │
                                      │       data.W, (A) │
								      └───────────────────┘
*/

// OnCommandTyped registers a typed handler function for a specific command type.
func OnCommandTyped[T any](cr *CommandRouter, cmdType any, handler TypedCommandHandlerFunc[T]) {
	if cr.component == nil {
		panic("CommandRouter not initialized - call Init() first")
	}

	// Wrap the typed handler
	wrappedHandler := func(cmd UpdateCmd) {
		// Try to handle as global command first
		if cr.component.HandleGlobalUpdate(cmd) {
			return
		}

		// Cast the data and call the typed handler
		if data, ok := cmd.Data.(T); ok {
			handler(data)
		}
	}

	cr.component.RegisterCommandHandler(cmdType, wrappedHandler)
}

// OnCommandTypedRaw registers a typed handler that runs even for global commands.
func OnCommandTypedRaw[T any](cr *CommandRouter, cmdType any, handler TypedCommandHandlerFunc[T]) {
	if cr.component == nil {
		panic("CommandRouter not initialized - call Init() first")
	}

	// Wrap handler to always try global commands first, then call custom handler
	wrappedHandler := func(cmd UpdateCmd) {
		// Try global first (may not match)
		cr.component.HandleGlobalUpdate(cmd)

		// Cast the data and call the typed handler
		if data, ok := cmd.Data.(T); ok {
			handler(data)
		}
	}

	cr.component.RegisterCommandHandler(cmdType, wrappedHandler)
}
