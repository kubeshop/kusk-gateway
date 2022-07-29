## kusk completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(kusk completion bash)

To load completions for every new session, execute once:

#### Linux:

	kusk completion bash > /etc/bash_completion.d/kusk

#### macOS:

	kusk completion bash > /usr/local/etc/bash_completion.d/kusk

You will need to start a new shell for this setup to take effect.


```
kusk completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.kusk.yaml)
```

### SEE ALSO

* [kusk completion](kusk_completion.md)	 - Generate the autocompletion script for the specified shell

