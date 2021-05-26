# Email Reply Parser
Email Reply Parser is a Golang library to parse plain-text email replies and extract content

This library supports most email replies, signatures and locales and quoted replies.

This library is used at **small** scale within webRidge.

- Supports stripping quoted replies in top/bottom
- Strip email replies like On DATE, NAME <EMAIL> wrote:
- Removes signatures like Sent from my iPhone
- Detects signatures like
```
Met vriendelijke groeten,
Richard Lindhout
```

```
Karen The Green
Graphic Designer
Office
Tel: +44423423423423
Fax: +44234234234234
karen@webby.com
Street 2, City, Zeeland, 4694EG, NL
www.thing.com

The content of this email is confidential and intended for the recipient specified in message only. It is strictly forbidden to share any part of this message with any third party, without a written consent of the sender. If you received this message by mistake, please reply to this message and follow with its deletion, so that we can ensure such a mistake does not occur in the future.

```

```
-Abhishek Kona


_______________________________________________
riak-users mailing list
riak-users@lists.basho.com
http://lists.basho.com/mailman/listinfo/riak-users_lists.basho.com
```

We try to support the following languages
- Dutch (tested)
- English (tested)
- French
- Polish
- German
- Portuguese
- Norwegian
- Swedish
- Danish
- Vietnamese


## TODO
- Forwarded emails


Please add more tests for your language and use-cases 
