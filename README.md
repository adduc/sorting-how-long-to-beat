# Sorting through How Long to Beat data

I am interested in playing through high quality games that take a
relatively short amount of time to beat, but Steam and other platforms
don't provide a good way to sort through games by their playlength. To
accomplish this, I found a dataset through Kaggle containing data from
[How Long to Beat](https://howlongtobeat.com/), a website that allows
users to submit their playtimes for games. I will use this data to
create a list of games that are highly rated and take a relatively short
amount of time to beat.

## Status

The import script works to import the data from the jsonlines file into
an SQLite database. I'm considering creating a web interface to allow
users to sort through the data.

The data is a bit old, and may necessitate updating to get the most
recent data.

## Queries

Find games that take up to 10 hours to beat on the PS1:

```sql
SELECT 
  games.name,
  games.release_year,
  games.review_score,
  game_platforms.time_to_beat/60 AS hours_to_beat
FROM game_platforms 
JOIN games ON games.id = game_platforms.game_id
JOIN platforms ON platforms.id = game_platforms.platform_id
WHERE platforms.name = 'PlayStation'
AND time_to_beat BETWEEN 1 AND 10*60
AND review_score > 60
ORDER BY review_score DESC
```


## Links:

- [Kaggle Dataset](https://www.kaggle.com/datasets/baraazaid/how-long-to-beat-video-games)