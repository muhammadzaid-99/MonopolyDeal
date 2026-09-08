# cashdeal

cashdeal is a multiplayer server for the card game Monopoly Deal. It lets people
create a room, share a short code with friends, and play a full game together
over a WebSocket connection.

The server keeps the whole game in memory and runs the rules itself. Players send
what they want to do, such as playing a card or ending their turn, and the server
works out what actually happens and sends the updated game back to everyone. The
game rules it follows are the standard ones, taken from
[monopolydealrules.com](https://monopolydealrules.com/).

There is a web client for this server in a separate repository,
[MonopolyDealClient](https://github.com/muhammadzaid-99/MonopolyDealClient), if
you want something to play on.

## What it does

**Runs many games at once.** Every room is independent, so games do not interfere
with each other and a problem in one room never affects the others.

**Handles the whole game.** All 106 cards are implemented, including properties,
wildcards, money, rent, houses and hotels, and every action card. The server
tracks turns, hands, banks and property sets for two to five players, and knows
when someone has won.

**Keeps information private.** You only ever receive your own hand. Of the other
players you see what you would see at a real table: their bank, their properties,
and how many cards they are holding, but never which cards those are.

**Handles cards that need a reply.** Some cards cannot finish on their own. If
someone plays Rent, the other players have to pay, and any of them might answer
with Just Say No first. The server pauses the turn, asks each player what they
want to do, collects the payments, and only then carries on.

**Lets you organise your properties.** Property sets are piles that you build
yourself. You can start a new pile, move cards between piles, and change the
colour of a pile built around a wildcard. Tidying up like this is free and does
not use one of your plays.

**Survives players dropping out.** Phones lock and networks change. If a player
disconnects, their seat and cards stay exactly where they were, and they can come
straight back to them.

## How a turn works

| Step | What happens |
| --- | --- |
| Draw | Draw 2 cards at the start of your turn, or 5 if your hand is empty |
| Play | Play up to 3 cards, into your bank, onto your properties, or as an action |
| Tidy | Rearrange your property piles as much as you like, at no cost |
| Discard | If you are holding more than 7 cards, discard down to 7 |
| End | Pass the turn to the next player |

The first player to complete three full property sets wins. Every play is checked
before it is allowed. You cannot play before drawing, cannot play a fourth card,
cannot end your turn while holding too many cards, and cannot act on someone
else's turn. If a card turns out to be unplayable in the current position, it
goes back into your hand and your play is not spent.

## Properties, sets and rent

Property sets here are real piles rather than something worked out from the
colours you happen to own. You create a pile, put cards into it, and the server
keeps that pile up to date as it changes. This matters because of wildcards: a
two colour wildcard can sit in either pile, and the same card can mean different
things depending on where you put it. Making the pile the thing you own removes
that ambiguity, and lets you rearrange your board without the server having to
guess your intent.

Each pile knows its colour, how many cards a full set of that colour needs, and
what it is currently worth. Rent is recalculated whenever the pile changes.

| Pile state | Rent |
| --- | --- |
| Not yet complete | The rent for the number of cards it holds |
| Complete | The full set rent |
| Complete with a house | Set rent plus 3 |
| Complete with a hotel | Set rent plus 7 |

Houses and hotels follow the usual restrictions. A house only goes on a completed
set, a hotel only goes on a set that already has a house, and neither can be
placed on railroads or utilities. A completed set will not accept further
property cards. Cards played to your board without a destination sit to one side
as loose cards until you file them into a pile, and empty piles are cleared away
at the end of your turn.

## Paying

When you owe someone money, you choose which cards to hand over, one at a time,
from your bank or from your properties. The server adds up what you have paid so
far and tells you when the debt is settled. As in the real game there is no
change given, so overpaying is your own problem.

If you genuinely cannot cover the amount, you hand over what you can and the debt
is closed there. Cards you pay with keep their nature: money and action cards go
into the other player's bank, properties go onto their board.

Just Say No can be used to cancel the demand outright, but only before you have
started paying. Once the first card is handed over the offer is considered
accepted.

## Cards that need a reply

Most cards resolve the moment they are played. Six do not, because they involve
another player who has to be given a chance to respond. These are handled as
small state machines that hold the game until they finish.

| Card | What the server has to arrange |
| --- | --- |
| Deal Breaker | Pick an opponent and one of their completed sets, then let them refuse |
| Sly Deal | Pick a single property from an opponent's incomplete set, then let them refuse |
| Forced Deal | Pick a property from each side to swap, then let them refuse |
| Debt Collector | Pick one opponent, then collect 5 from them |
| It's My Birthday | Collect 2 from every other player |
| Rent | Choose a set and, depending on the card, one opponent or all of them |

Each of these moves through the same shape: a selection step, a chance to react,
and then payment. While one is open, everything a player sends is routed into it
instead of the normal turn handling, so nobody can wander off and play a card
while a payment is outstanding. Each player's debt is tracked separately, which
is what allows It's My Birthday and a two colour Rent to collect from several
people at once, in whatever order they happen to answer.

Two rules fall out of this design rather than needing special cases of their own.
Just Say No is dealt with inside whichever action it is answering, so playing one
against another is simply the same step happening twice, and the chain can go on
as long as the players have the cards for it. Double The Rent is not an effect by
itself: it reaches into the rent that is already waiting to be paid and doubles
it, which is exactly what the printed card describes.

## Rooms and their lifecycle

When someone creates a room they get a six digit code to share, checked against
the rooms already running so two games never end up with the same one. Up to five
players join with that code, everyone marks themselves ready, and any player can
then start the game.

Rooms look after themselves. There is no cleanup job sweeping the server, no
manual teardown, and nothing left behind when a game ends.

| Situation | What happens |
| --- | --- |
| Somebody is still connected | The room stays |
| Everyone has gone and no players are seated | Deleted after 2 minutes |
| Everyone has gone but players are still seated | Deleted after 20 minutes |
| Anyone joins or reconnects | The countdown is cancelled |

The two waits are deliberate. An empty lobby that nobody joined is worth
forgetting quickly, but a real game where everyone briefly lost signal deserves a
generous window before it is thrown away.

Leaving works differently before and after the game starts. In the lobby you can
leave freely and your seat is released. Once play has begun, leaving is treated
as a disconnection: you are marked absent, but your hand, bank and properties
stay on the table so the game is still there if you come back.

## Players and reconnecting

The server hands out identity rather than trusting the client for it. On your
first message it generates a player ID, records which connection and address it
belongs to, and sends it back. Your client keeps it and includes it from then on,
which is what stops one player from acting as another by simply claiming their
ID.

Because your identity is not the socket, losing the socket costs you nothing. If
a message arrives for a known player on a new connection, the server recognises
that the player has come back, points their seat at the new connection, and tells
the room. It also notices when the address behind a player changes, which is what
happens when a phone moves from wifi to mobile data, and reports that as a
reconnection rather than a stranger.

The last thing the server asked of each player is remembered as well. If you drop
out halfway through paying rent, you return to that same request instead of a
board that looks finished but will not let you do anything.

## How it is built

The server is written in Go and uses [gws](https://github.com/lxzan/gws) for
WebSocket connections. There is no database. Everything lives in memory for as
long as the room does, which is the right trade for a game that is worthless once
it is over.

### One room, one goroutine

Each connection is read on its own goroutine, so two players in the same game can
easily be sending messages at the same instant. The obvious answer is to lock the
game state everywhere it is touched, which gets unpleasant quickly when a single
action moves cards between four different players.

Instead, every room runs its own goroutine with its own queue of events.
Connections do not touch game state at all. They put a message on the room's
queue and move on, and the room takes them off one at a time in the order they
arrived. Only one thing is ever happening to a game, so the game logic can be
written as ordinary sequential code and read as if it were single player. Where a
caller does need an answer, such as finding out whether a join was accepted, it
sends a reply channel along with the event and waits for the room to answer.

This also draws a clean line around failure. If a room ever panics it recovers,
logs what happened, and hands its own ID to a cleanup worker that removes it and
releases its memory. One broken game disappears. Every other game on the server
carries on without noticing.

### Changing state safely

Moving a card between players is not a single step. The server has to find the
card, check that taking it is legal, and only then remove it from where it was.
Doing that in one pass risks pulling a card out and discovering afterwards that
the move was not allowed.

So a lookup returns the card along with a function that commits the removal. The
caller inspects what it found, and only calls that function once it is sure. The
commit can only run once, and it also tidies up after itself by clearing the pile
if it is now empty and recalculating rent. The same idea applies to playing a
card: it leaves your hand first, and if the play turns out to be illegal it is
put straight back and the play is not counted.

### The deck

The deck is built once per game from the individual card groups and shuffled. The
106 cards are numbered from 1 upwards as the deck is assembled, and that number
is the only thing the client ever needs to send to refer to a card. Zero is never
used, so a missing or malformed card ID cannot accidentally mean a real card.

When the draw pile runs out, the discard pile is shuffled back into it, except
for the card on top, which stays visible so play can continue from a known
position.

## Talking to the server

Everything in both directions is JSON with a type on it, over one connection per
player.

| The client sends | For |
| --- | --- |
| `create-room`, `join-room` | Getting into a game |
| `change-ready-state`, `start-game` | The lobby |
| `draw-cards`, `play-card`, `end-turn` | Taking a turn |
| `arrange-property`, `create-pile`, `change-pile-color` | Organising your board |
| `discard-card` | Getting back to seven cards |

While a card is waiting on a reply, the same messages carry the answer to it, so
the client does not need a separate vocabulary for responding to Rent or choosing
who to take a set from.

| The server sends | What it contains |
| --- | --- |
| Your hand | Only ever your own cards |
| The player list | Everyone's name, bank, properties and hand size |
| Game info | Whose turn it is, plays left, the top of the discard pile, recent events |
| Your prompt | What you are expected to do right now |

Any change sends all of these together rather than a description of what changed.
It costs a little more traffic and removes a whole category of bug, because the
client can never drift out of step with the server. It simply draws whatever
arrived last, and a client that missed a message is fixed by the next one.

## Running it

You need Go 1.24 or newer.

```sh
go run .
```

The server listens on the port in the `PORT` environment variable, or 8080 if
that is not set. On Windows, `run.bat` builds and starts it in one step.

## Project layout

| Files | What is in them |
| --- | --- |
| `main.go` | Starting the server |
| `gws-main.go`, `gws-game.go` | Connections, player identity, and the room event loop |
| `room.go`, `room-manager.go` | Rooms, turns, and sending game state to players |
| `action.go`, `action-resolver.go` | Action cards and the ones that need a reply |
| `models/cards` | The cards and the deck |
| `models/game` | Game state, players, property piles and rent |
