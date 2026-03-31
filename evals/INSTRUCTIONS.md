# Instructions

Some tips for validating label offsets:
1. Open the eval file in vim
2. Use e.g. 112<space> to jump to byte 112 in the file
3. Use `g CTRL+G` to make vim show the byte character offset in the file

Either manually eyeball the rest of the document for patterns Claude might have
missed when generating labels, or run the eval suite and check for
False Positives, indicating something was flagged that wasn't mentioned in the
labels file.
