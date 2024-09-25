import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const protocols = ['http', 'https', 'ftp', 'sftp'];
const tlds = ['com', 'org', 'net', 'io', 'edu', 'gov'];

export function randomString(length) {
    const charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += charset[randomIntBetween(0, charset.length - 1)];
    }
    return result;
}

export function randomUrl(maxLength = 400) {
    let url = protocols[randomIntBetween(0, protocols.length - 1)] + '://';
    url += randomString(randomIntBetween(1, 20)) + '.' + tlds[randomIntBetween(0, tlds.length - 1)];
    
    const pathSegments = randomIntBetween(0, 5);
    for (let i = 0; i < pathSegments; i++) {
        url += '/' + randomString(randomIntBetween(1, 15));
    }

    const queryParams = randomIntBetween(0, 5);
    if (queryParams > 0) {
        url += '?';
        for (let i = 0; i < queryParams; i++) {
            if (i > 0) url += '&';
            url += randomString(randomIntBetween(1, 10)) + '=' + randomString(randomIntBetween(1, 20));
        }
    }

    if (randomIntBetween(0, 1) === 1) {
        url += '#' + randomString(randomIntBetween(1, 10));
    }

    if (url.length > maxLength) {
        url = url.substring(0, maxLength);
    }

    return url;
}