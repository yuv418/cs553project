A good stable world seed: `8525333463046388971`


```
INPUT_CSV=input_seeds/8525333463046388971.csv GAME_URL=http://localhost:5173 poetry run python -i src/input_simulator.py  --origin-to-force-quic-on=localhost:4433,localhost:4434 --ignore-certificate-errors-spki-list=tEQISBkCOe6IzTBZxDmUHotYwprs5lXNxPUiM71ySwo=
```
INPUT_CSV=input_seeds/<SEED>.csv GAME_URL=<URL> poetry run python -i src/input_simulator.py  --origin-to-force-quic-on=<URLS> --ignore-certificate-errors-spki-list=<LIST>
