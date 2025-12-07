# Backend Test

> Source: https://github.com/elotusteam/challenges/blob/main/backend.md

- Total time: 50 hours (please no more!)

## Data Structures And Algorithms

> Time: 2 hours
> Please commit and push the code of this round after 2 hours of receiving the test, then you can start the Hackathon round

1. [Gray Code](/code-challenges/gray-code.md)

2. [Sum of Distances in Tree](/code-challenges/sum-of-distances-in-tree.md)

3. [Maximum Length of Repeated Subarray](/code-challenges/maximum-length-of-repeated-subarray.md)

## Hackathon

> Time: 48 hours

Write a server with 2 features:

1. Write a simple authentication using JWT token (with HS2256)

   - Register (just need username and password)
   - Login
   - Revoke token by time

> Please note: you do not need implement full features OAuth2, just follow our requirements

2. Write an API handler a form upload file

   - A file upload field named "data" (ie, `<input type="file"/>`) that uploads
     a file that the user selects (please do not waste time trying to make the
     HTML form pretty -- we don't need HTML developers, we need Backend
     developers)

   - The form above should POST data to the `/upload` handler, which should write
   the received file data to a temporary file in /tmp.

   - Before accepting any data, you should check authorization (token generate via feature 1) is valid and the content type of the
   uploaded file is an image. If the submission is bad, please return an error. Images larger than 8
   megabytes should also be rejected.

   - Write the image metadata (content type, size, etc) to a database of your
   choice, including all relevant HTTP information.

## Notes

- Use Go.

- With The Hackathon, feel free to use any open source packages that you want -- we're not looking
  for developers to constantly reinvent the wheel, but for developers who are
  able to make use of existing packages. This is a skills based assessment,
  not an attempt to assess whether or not you can write a form upload
  library.
  
- Do a short 1 or 2 paragraph write up with information on building/running the application if relevant / necessary or non-standard.

- It is not a problem if you are not able to complete all of the above in a
  limited time. Part of this assessment is to determine what progress you are
  able to make in that timeframe.

- You should commit and push code for each feature, it will prove you know how to use git. Please do not commit all code in one commit

- When you are done, please push all work to a Git repository, and send us
  the repository url. Please **make sure the repository is public**.

- Bonus points:

  - Properly comment your code and the use of environment variables or flags
    where possible.
  - Properly write a readme on how to get your code up and run.
  - Use as few Open Source libs as possible. That shows how strong you are at
    your choosen language.

- If you make it to the next round, we will talk about this project, and you
  will be expected to explain your design, decisions,... So, after submit the
  code, take some time to think about what can be asked.
