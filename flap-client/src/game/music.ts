import type { PlayMusicResp } from '../protos/music/music_pb';
import { logReceiveTime } from './latencyLogger';

// Initiatlize stuff

const context = new AudioContext();


// Copied from https://stackoverflow.com/questions/37228285/uint8array-to-arraybuffer
// There's really no point in rewriting this myself.
function typedArrayToBuffer(array: Uint8Array): ArrayBuffer {
    return array.buffer.slice(array.byteOffset, array.byteLength + array.byteOffset)
}

// decodeAudioData
// https://stackoverflow.com/questions/24151121/how-to-play-wav-audio-byte-array-via-javascript-html5
export function playSound(_: string, resp: PlayMusicResp) { 
    logReceiveTime('audio');
    context.decodeAudioData(typedArrayToBuffer(resp.audioPayload), (retBuf) => {
        let src = context.createBufferSource()

        src.buffer = retBuf
        src.connect(context.destination)
        src.start(0)
    })
}
