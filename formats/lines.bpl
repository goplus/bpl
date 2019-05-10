line = {
	let line, _ = BPL_IN.readString('\n')
	do printf("%s", line)
	return line
}

doc = *[line] dump
