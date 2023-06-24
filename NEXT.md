## TODO
- Write dockerized integration tests. Use docker-compose to start both
  the client and server. Fire events and assert the saved sessions. Make it
  tolerate a seconds time difference for each session.

- Add a .code-harvest config to repos that includes some information like type of project (infra/frontend/backend)

- Should the `cron` folder in cmd be changed to something like "aggregation"?

- Maybe storage should not house the New functions for memory/mongo/file

- Make the disk storage take a path? Could create a temporary directory during tests
