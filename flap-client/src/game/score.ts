import { ScoreService } from '../protos/score/score_pb';
import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";

export const scoreTransport = createConnectTransport({
    baseUrl: import.meta.env.VITE_SCORE_SERVICE_URL,
    useBinaryFormat: true,
});

export const scoreClient = createClient(ScoreService, scoreTransport);
