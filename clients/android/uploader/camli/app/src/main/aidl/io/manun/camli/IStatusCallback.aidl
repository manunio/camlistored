// IStatusCallback.aidl
package io.manun.camli;

interface IStatusCallback {
    oneway void logToClient(String stuff);
    oneway void onUploadStatusChange(boolean uploading);
}