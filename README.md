[![Netlify Status](https://api.netlify.com/api/v1/badges/cb0773c9-7d6a-48b9-bdbb-428f6e4359ae/deploy-status)](https://app.netlify.com/sites/rococo-genie-d49178/deploys)

# PureMOOt

This is a version of permoot (A discord bot in the Unofficial MOP 2022 Server)
for Lemonade Shop (Unofficial MOP 2023 Server) with the following functionality:

1. Let people anonymously broadcast messages
2. Issue "pureMOOtations" to encourage people to start private conversations
   and make new friends.
3. Administer a version of the (in)famous [AoPS reaper](https://aops.com/reaper)

# Setup

1. The bot is invited
2. Create a role called "cow"; members who have this role will be paired up in a
   puremootation.
3. Users of the `/puremoot` command must either have the administrator or
   have both the "manage server" and "view audit log" permission. These users
   are also capable of using the `/reapergame init` and `/reapergame cancel`
   commands

# Commands

- `/broadcast [message]` - anonymously broadcast a message.
- `/puremoot [day number]` (Admins only) - generate a puremootation.
- `/reapergame [subcommand]` - All operations related to reaper (certain
  commands are Admin only)
- `/reap` - harvest a pear.

# Acknowledgements

- Advaith Avadhanam, Lemonade Shop Dictator
- Emily Yu, for designing the logo
