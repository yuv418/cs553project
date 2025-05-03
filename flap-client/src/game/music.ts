import type { PlayMusicResp } from '../protos/music/music_pb';

export function playSound(_: PlayMusicResp) {
    console.log("Got audio from music stream")
}
