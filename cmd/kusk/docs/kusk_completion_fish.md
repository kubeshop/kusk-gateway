## kusk completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	kusk completion fish | source

To load completions for every new session, execute once:

	kusk completion fish > ~/.config/fish/completions/kusk.fish

You will need to start a new shell for this setup to take effect.


```
kusk completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.kusk.yaml)
```

### SEE ALSO

* [kusk completion](kusk_completion.md)	 - Generate the autocompletion script for the specified shell

