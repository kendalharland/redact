# redact

A CLI tool that redacts sensitive data from text using pattern matching and LLM-based detection.

## Quickstart

```bash
# Build the binary
make redact

# Set your Anthropic API key (required for person name and obfuscated data detection)
export ANTHROPIC_API_KEY=your-key-here

# Redact sensitive data from a file
./bin/redact input.txt

# Redact from stdin
echo "Contact john.doe@example.com at 555-123-4567" | ./bin/redact
```

## Supported Data Types

**Pattern-based (regex):**
- Credit card numbers
- Email addresses
- IP addresses (IPv4 and IPv6)
- MAC addresses
- Phone numbers

**LLM-based (requires API key):**
- Person names
- Obfuscated data (e.g., "one-two-three" for "123")

## CLI Usage

### Basic redaction

```bash
# Redact a file, output to stdout
./bin/redact document.txt

# Redact from stdin
cat document.txt | ./bin/redact
```

### Bleep mode

Replace sensitive data with `***` instead of typed placeholders:

```bash
./bin/redact --bleep document.txt
```

### Exclude specific types

Skip certain data types during redaction:

```bash
# Skip email and phone detection
./bin/redact --exclude=EMAIL_ADDRESS,PHONE_NUMBER document.txt
```

### Export classification map and labels

```bash
# Save the classification map (what was found)
./bin/redact --output-cmap=classmap.json document.txt

# Save the labels (where each item was found)
./bin/redact --output-labels=labels.json document.txt
```

### Use a different Claude model

```bash
./bin/redact --model=claude-haiku-4-5 document.txt
```

## Running Evaluations

The tool includes an evaluation system to measure precision and recall:

```bash
# Run all evaluations in the evals/ directory
./bin/redact evals

# Run specific evaluation files
./bin/redact evals evals/basic.txt evals/mixed-pii.txt

# Output metrics as JSON
./bin/redact evals --json-output=metrics.json
```

## Example

Input:
```
Meeting with John Smith tomorrow.
His email is john.smith@acme.com and phone is (555) 123-4567.
```

Output:
```
Meeting with <PERSON_1> tomorrow.
His email is <EMAIL_ADDRESS_1> and phone is <PHONE_NUMBER_1>.
```

With `--bleep`:
```
Meeting with *** tomorrow.
His email is *** and phone is ***.
```
