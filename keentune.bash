# keentune(1) completion   

_keentune()  
{
    COMPREPLY=()
    local commands="--help param profile sensitize --version -v"
    local cur=${COMP_WORDS[COMP_CWORD]};
    local cmd=${COMP_WORDS[COMP_CWORD-1]};
    case $cmd in
    'keentune')
          COMPREPLY=( $(compgen -W '$commands' -- $cur) ) ;;
    'param')
          COMPREPLY=( $(compgen -W 'delete dump jobs list rollback stop tune' -- $cur) ) ;;
    'tune' | 'dump')
          COMPREPLY=( $(compgen -W '--job' -- $cur) ) ;;
    'profile')
          COMPREPLY=( $(compgen -W 'delete generate info list rollback set' -- $cur) ) ;;
    'generate' | 'info')
          COMPREPLY=( $(compgen -W '--name' -- $cur) ) ;;	  
    'sensitize')
          COMPREPLY=( $(compgen -W 'delete jobs stop train' -- $cur) ) ;;
    'train')
          COMPREPLY=( $(compgen -W '--data' -- $cur) ) ;;		  
    'set')
          COMPREPLY=( $(compgen -W '--group1' -- $cur) ) ;; 
    esac

    #profile set command
    if [[ "${COMP_WORDS[2]}" == "set" && "${COMP_WORDS[3]}" == "--group1" ]]; then
	    local pro=($(pwd))
	    cd /etc/keentune/profile
	    compopt -o nospace
	    COMPREPLY=($(compgen -d -f -- $cur))
	    cd $pro
    fi

    #param/profile/sensitize delete command
    if [[ "$cmd" == "delete" && ( "${COMP_WORDS[1]}" == "sensitize" || "${COMP_WORDS[1]}" == "param" )]]; then
	    COMPREPLY=( $(compgen -W '--job' -- $cur) )
    elif [[ "$cmd" == "delete" && "${COMP_WORDS[1]}" == "profile" ]]; then
	    COMPREPLY=( $(compgen -W '--name' -- $cur) )
    fi

    return 0
} 
complete -F _keentune keentune
