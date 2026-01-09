# exarp-go bash completion
_exarp_go() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="-tool -args -list -test -i -completion"

    case "${prev}" in
        -tool|--tool)
            COMPREPLY=($(compgen -W "task_workflow recommend setup_hooks task_discovery ollama session prompt_tracking infer_session_mode list_models context analyze_alignment automation lint check_attribution security tool_catalog add_external_tool_hints git_tools server_status health memory_maint testing mlx context_budget generate_config memory report workflow_mode estimation task_analysis" -- "${cur}"))
            return 0
            ;;
        -test|--test)
            COMPREPLY=($(compgen -W "task_workflow recommend setup_hooks task_discovery ollama session prompt_tracking infer_session_mode list_models context analyze_alignment automation lint check_attribution security tool_catalog add_external_tool_hints git_tools server_status health memory_maint testing mlx context_budget generate_config memory report workflow_mode estimation task_analysis" -- "${cur}"))
            return 0
            ;;
        -completion|--completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "${cur}"))
            return 0
            ;;
        *)
            if [[ ${cur} == -* ]] ; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            fi
            ;;
    esac
}
complete -F _exarp_go exarp-go
