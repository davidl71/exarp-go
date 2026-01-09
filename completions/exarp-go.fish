# exarp-go fish completion

complete -c exarp-go -n '__fish_use_subcommand' -s t -l tool -d 'Tool name to execute' -xa 'setup_hooks report task_workflow ollama analyze_alignment infer_session_mode prompt_tracking list_models add_external_tool_hints memory lint estimation server_status memory_maint task_analysis testing workflow_mode session mlx context health recommend generate_config git_tools context_budget security task_discovery automation tool_catalog check_attribution'
complete -c exarp-go -n '__fish_use_subcommand' -l args -d 'Tool arguments as JSON'
complete -c exarp-go -n '__fish_use_subcommand' -s l -l list -d 'List all available tools'
complete -c exarp-go -n '__fish_use_subcommand' -l test -d 'Test a tool with example arguments' -xa 'setup_hooks report task_workflow ollama analyze_alignment infer_session_mode prompt_tracking list_models add_external_tool_hints memory lint estimation server_status memory_maint task_analysis testing workflow_mode session mlx context health recommend generate_config git_tools context_budget security task_discovery automation tool_catalog check_attribution'
complete -c exarp-go -n '__fish_use_subcommand' -s i -l interactive -d 'Interactive mode'
complete -c exarp-go -n '__fish_use_subcommand' -l completion -d 'Generate shell completion script' -xa 'bash zsh fish'
