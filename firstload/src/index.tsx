import {
    DebugOptions,
    Highlight,
    HighlightClassOptions,
} from '../../client/src/index';
import packageJson from '../package.json';
import { listenToChromeExtensionMessage } from './browserExtension/extensionListener';
import { SessionDetails } from './types/types';

export type HighlightOptions = {
    // a 'true' value defaults to only loggin api interactions.
    debug?: boolean | DebugOptions;
    scriptUrl?: string;
    backendUrl?: string;
    manualStart?: boolean;
    disableNetworkRecording?: boolean;
    /**
     * This enables recording XMLHttpRequest and Fetch headers and bodies.
     * disableNetworkRecording needs to be false.
     * Defaults to false.
     */
    enableNetworkHeadersAndBodyRecording?: boolean;
    disableConsoleRecording?: boolean;
    enableSegmentIntegration?: boolean;
    /**
     * The environment your application is running in.
     * This is useful to distinguish whether your session was recorded on localhost or in production.
     */
    environment?: 'development' | 'staging' | 'production' | string;
    /**
     * The version of your application.
     * This is commonly a Git hash or a semantic version.
     */
    version?: string;
    /**
     * Enabling this will disable recording of text data on the page. This is useful if you do not want to record personally identifiable information and don't want to manually annotate your code with the class name "highlight-block".
     * @example
     * // Text will be randomized. Instead of seeing "Hello World" in a recording, you will see "1fds1 j59a0".
     * @see {@link https://docs.highlight.run/docs/privacy} for more information.
     */
    enableStrictPrivacy?: boolean;
};

const HighlightWarning = (context: string, msg: any) => {
    console.warn(`Highlight Warning: (${context}): `, msg);
};

export interface HighlightPublicInterface {
    init: (orgID: number | string, debug?: HighlightOptions) => void;
    identify: (identify: string, obj: any) => void;
    track: (event: string, obj: any) => void;
    /**
     * @deprecated with replacement by `consumeError` for an in-app stacktrace.
     */
    error: (message: string, payload?: { [key: string]: string }) => void;
    /**
     * Calling this method will report an error in Highlight and map it to the current session being recorded.
     * A common use case for `H.error` is calling it right outside of an error boundary.
     * @see {@link https://docs.highlight.run/docs/error-handling} for more information.
     */
    consumeError: (
        error: Error,
        message?: string,
        payload?: { [key: string]: string }
    ) => void;
    getSessionURL: () => Promise<string>;
    getSessionDetails: () => Promise<SessionDetails>;
    start: () => void;
    /** Stops the session and error recording. */
    stop: () => void;
    onHighlightReady: (func: () => void) => void;
    options: HighlightOptions | undefined;
}

interface HighlightWindow extends Window {
    Highlight: new (options?: HighlightClassOptions) => Highlight;
    H: HighlightPublicInterface;
}

const HIGHLIGHT_URL = 'app.highlight.run';

declare var window: HighlightWindow;

var script: HTMLScriptElement;
var highlight_obj: Highlight;
export const H: HighlightPublicInterface = {
    options: undefined,
    init: (orgID: number | string, options?: HighlightOptions) => {
        try {
            H.options = options;

            // Don't run init when called outside of the browser.
            if (
                typeof window === 'undefined' ||
                typeof document === 'undefined'
            ) {
                return;
            }

            script = document.createElement('script');
            var scriptSrc = options?.scriptUrl
                ? options.scriptUrl
                : 'https://static.highlight.run/index.js';
            script.setAttribute(
                'src',
                scriptSrc + '?' + new Date().getMilliseconds()
            );
            script.setAttribute('type', 'text/javascript');
            document.getElementsByTagName('head')[0].appendChild(script);
            script.addEventListener('load', () => {
                highlight_obj = new window.Highlight({
                    organizationID: orgID,
                    debug: options?.debug,
                    backendUrl: options?.backendUrl,
                    disableNetworkRecording: options?.disableNetworkRecording,
                    enableNetworkHeadersAndBodyRecording:
                        options?.enableNetworkHeadersAndBodyRecording,
                    disableConsoleRecording: options?.disableConsoleRecording,
                    enableSegmentIntegration: options?.enableSegmentIntegration,
                    enableStrictPrivacy: options?.enableStrictPrivacy || false,
                    firstloadVersion: packageJson['version'],
                    environment: options?.environment || 'production',
                    appVersion: options?.version,
                });
                if (!options?.manualStart) {
                    highlight_obj.initialize(orgID);
                }
            });
        } catch (e) {
            HighlightWarning('init', e);
        }
    },
    consumeError: (
        error: Error,
        message?: string,
        payload?: { [key: string]: string }
    ) => {
        try {
            H.onHighlightReady(() =>
                highlight_obj.consumeCustomError(
                    error,
                    message,
                    JSON.stringify(payload)
                )
            );
        } catch (e) {
            HighlightWarning('error', e);
        }
    },
    error: (message: string, payload?: { [key: string]: string }) => {
        try {
            H.onHighlightReady(() =>
                highlight_obj.pushCustomError(message, JSON.stringify(payload))
            );
        } catch (e) {
            HighlightWarning('error', e);
        }
    },
    track: (event: string, obj: any) => {
        try {
            H.onHighlightReady(() =>
                highlight_obj.addProperties({ ...obj, event: event })
            );
        } catch (e) {
            HighlightWarning('track', e);
        }
    },
    start: () => {
        try {
            if (highlight_obj?.state === 'Recording') {
                console.warn(
                    'You cannot called `start()` again. The session is already being recorded.'
                );
                return;
            }
            if (H.options?.manualStart) {
                var interval = setInterval(function () {
                    if (highlight_obj) {
                        clearInterval(interval);
                        highlight_obj.initialize();
                    }
                }, 200);
            } else {
                console.warn(
                    "Highlight Error: Can't call `start()` without setting `manualStart` option in `H.init`"
                );
            }
        } catch (e) {
            HighlightWarning('start', e);
        }
    },
    stop: () => {
        try {
            H.onHighlightReady(() => highlight_obj.stopRecording(true));
        } catch (e) {
            HighlightWarning('stop', e);
        }
    },
    identify: (identifier: string, obj: any) => {
        try {
            H.onHighlightReady(() => highlight_obj.identify(identifier, obj));
        } catch (e) {
            HighlightWarning('identify', e);
        }
    },
    getSessionURL: () => {
        return new Promise<string>((resolve, reject) => {
            H.onHighlightReady(() => {
                const orgID = highlight_obj.organizationID;
                const sessionID = highlight_obj.sessionData.sessionID;
                if (orgID && sessionID) {
                    const res = `${HIGHLIGHT_URL}/${orgID}/sessions/${sessionID}`;
                    resolve(res);
                } else {
                    reject(new Error('org ID or session ID is empty'));
                }
            });
        });
    },
    getSessionDetails: () => {
        return new Promise<SessionDetails>((resolve, reject) => {
            H.onHighlightReady(() => {
                const orgID = highlight_obj.organizationID;
                const sessionID = highlight_obj.sessionData.sessionID;
                if (orgID && sessionID) {
                    const currentSessionTimestamp = highlight_obj.getCurrentSessionTimestamp();
                    const now = new Date().getTime();

                    const baseUrl = `https://${HIGHLIGHT_URL}/${orgID}/sessions/${sessionID}`;
                    const url = new URL(baseUrl);

                    const urlWithTimestamp = new URL(baseUrl);
                    urlWithTimestamp.searchParams.set(
                        'ts',
                        // The delta between when the session recording started and now.
                        ((now - currentSessionTimestamp) / 1000).toString()
                    );

                    resolve({
                        url: url.toString(),
                        urlWithTimestamp: urlWithTimestamp.toString(),
                    });
                } else {
                    reject(new Error('org ID or session ID is empty'));
                }
            });
        });
    },
    onHighlightReady: (func: () => void) => {
        try {
            if (highlight_obj && highlight_obj.ready) {
                func();
            } else {
                var interval = setInterval(function () {
                    if (highlight_obj && highlight_obj.ready) {
                        clearInterval(interval);
                        func();
                    }
                }, 200);
            }
        } catch (e) {
            HighlightWarning('onHighlightReady', e);
        }
    },
};

if (typeof window !== 'undefined') {
    window.H = H;
}

listenToChromeExtensionMessage();
