# MefistoNotifications(r)

Notifications from dofamin.org main page.

## Running

You need PostgreSQL database installed.

```bash
~$ export MEF_CHATID="..."  # telegram ChatID
~$ export MEF_TGTOKEN="..."  # Bot token can be obtained by t.me/BotFather
~$ export MEF_DSN="postgresql://user:password@hostname:5432/database_name"  # URL for PostgreSQL connection.
~$ ./mefnotify
```
