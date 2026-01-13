# PR naming convention
Please use the following naming convention for your pull requests:

```
[emoji] [type]([scope]): [short description]
```

Where:
- `[emoji] [type]` is one of the following:
  - `✨ feat` for new features
  - `🐞 fix` for fixes
  - `🍒 pick` for picking changes from other branches
  - `♻️ refacto` for refactoring
  - `🧹 chore` for maintenance tasks
  - `🧪 test` for adding or updating tests
  - `📝 docs` for documentation changes
- `[scope]` component that specifies the buisness project affected by the changes, for example: [`bronze`, `silver`, `gold`, `core`].
- `[short description]` is a brief summary of the changes made in the pull request, it must be concise and descriptive.