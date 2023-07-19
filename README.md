# PureMOOt

Tea Shop (Unofficial MOP 2022 Server) has a discord bot named perMOOt that is
moderately useful but extremely annoying. In particular, it instantly replies
"carefully" to any message starting with "how" and will change a person's
nickname every time there is an "im" in their message (E.g. "somet**im**e soon",
"hi e soon, nice to meet you!")

This is a nerfed version of permoot for Lemonade Shop (Unofficial MOP 2023
Server) that aims to do exactly two things:

1. Let people anonymously broadcast messages
2. Issue "pureMOOtations" to encourage people to start private conversations
   and make new friends.

# Setup

1. The bot is invited
2. Create a role called "cow"; members who have this role will be paired up in a
   puremootation.
3. Users of the `/puremoot` command must either have the administrator
   permission or have a role that contains the substring "admin"

# Commands

- `/broadcast [message]` - anonymously broadcast a message.
- `/puremoot [day number]` (Admins only) - generate a puremootation.