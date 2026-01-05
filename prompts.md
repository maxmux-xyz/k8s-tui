Look at k8s-pod-manager-go-tui.md this is what we want to implement.
Look at ./docs to see what we've done for now.

Make a new plan to continue building our new TUI. Use ask question tool for any clarification you need
From the plan isolate what would be a next good self contained task.
We want to make methodical incremental progress, not so sloppy job all at once.
Make sure to plan contains tests as well.

Never enter plan mode.
We are making a plan and writing it to md file - but you should not enter plan mode.

Whilst making plan, you should:
1. Thoroughly explore the codebase to understand existing patterns
2. Identify similar features and architectural approaches
3. Consider multiple approaches and their trade-offs
4. Use AskUserQuestion if you need to clarify the approach
5. Design a concrete implementation strategy
f
--------------------------------------------------------------------------------

Look at ./docs/plans and pick up the next plan that is ready to be worked on andf has not been implemented yet.
Implement the plan.  Use ask question tool for any clarification you need.
When done, write implementation notes to docs/implemented.
And append "done" to the plan file name.

Then commit the changes to the repository.
`git add . && git commit -m "Implement plan X"`

Then add an md file in ./docs/human_test ()
- list some new things we should implement that the human should test.
- human on call will test those things and come back with feedback if anything is not working as expected.


--------------------------------------------------------------------------------

Look at ./docs to see what we've done for now.
Give me directions, as a user, on the following:
- What functionalities are implemented?
- How can i test the current implementation?

--------------------------------------------------------------------------------

Entered plan mode. You should now focus on exploring the codebase and designing an implementation approach.

In plan mode, you should:
1. Thoroughly explore the codebase to understand existing patterns
2. Identify similar features and architectural approaches
3. Consider multiple approaches and their trade-offs
4. Use AskUserQuestion if you need to clarify the approach
5. Design a concrete implementation strategy
6. When ready, use ExitPlanMode to present your plan for approval

Remember: DO NOT write or edit any files yet. This is a read-only exploration and planning phase.