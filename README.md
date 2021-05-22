# Email Reply Parser (WORK IN PROGRESS)
Email Reply Parser is a Golang library to parse plain-text email replies and extract content

This library supports most email replies, signatures and locales.

This library is used at **small** scale within webRidge.

- Strip email replies like On DATE, NAME <EMAIL> wrote:
- Removes signatures like Sent from my iPhone
- Removes signatures like Best wishes

We try to support the following languages
- Dutch (tested)
- English (tested)
- French
- Polish
- German
- Portuguese
- Norwegian
- Swedish, Danish
- Vietnamese


Please add more tests for your language and use-cases 
