# Workbench

Personal workbench for all the things I automate or don't want to forget. It's also a personal project to expand my Go knowledge.

## Installation

```sh
$ go install gitub.com/bdunn313/workbench@latest
```

## Usage

### Roll

Roll is a simple dice roller. It takes a dice notation and rolls the dice. It supports the following notation:

- `d6`: Roll a single 6-sided die
- `2d6`: Roll two 6-sided dice
- `d6+1`: Roll a single 6-sided die and add 1 to the result
- `2d6+1`: Roll two 6-sided dice and add 1 to the result
- `2d4+3d8+4+1d20+2`: Roll two 4-sided dice, three 8-sided dice, a single 20-sided die, and add 4 and 2 to the result

```sh
$ workbench roll 2d4+3d8+4+1d20+2
Rolling: 2d4+3d8+4+1d20+2

 2      2d4
30    3d8+4
15   1d20+2
------------
47
```

It also supports plain formatting so you can pipe the output

```sh
$ workbench roll 2d4+3d8+4+1d20+2 --plain | cat
47
```

**Note:** Right now it does not support negative numbers for modifiers. That is tracked [here](https://github.com/bdunn313/workbench/issues/1).
