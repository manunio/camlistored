package io.manun.camli;

import android.util.Log;

import org.apache.http.HttpResponse;
import org.apache.http.StatusLine;
import org.apache.http.auth.AuthScope;
import org.apache.http.auth.UsernamePasswordCredentials;
import org.apache.http.client.CredentialsProvider;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.impl.client.BasicCredentialsProvider;
import org.apache.http.impl.client.DefaultHttpClient;
import org.apache.http.message.BasicNameValuePair;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.IOException;
import java.io.UnsupportedEncodingException;
import java.util.ArrayList;
import java.util.LinkedList;
import java.util.List;
import java.util.concurrent.atomic.AtomicBoolean;

public class UploadThread extends Thread {
    private static final String TAG = UploadThread.class.getName();

    private final UploadService mService;
    private final HostPort mHostPort;
    private LinkedList<QueuedFile> mQueue;

    private final AtomicBoolean mStopRequested = new AtomicBoolean(false);

    private final DefaultHttpClient mUA = new DefaultHttpClient();

    public UploadThread(UploadService mService, HostPort mHostPort, String mPassword) {
        this.mService = mService;
        this.mHostPort = mHostPort;
        this.mQueue = mService.uploadQueue();

        CredentialsProvider creds = new BasicCredentialsProvider();
        creds.setCredentials(AuthScope.ANY, new UsernamePasswordCredentials("TOD-DUMMY-USER",
                mPassword));
        mUA.setCredentialsProvider(creds);
    }

    public void stopPlease() {
        mStopRequested.set(false);
    }

    @Override
    public void run() {
        if (!mHostPort.isValid())
            return;
        Log.d(TAG, "Running UploadThread for " + mHostPort);

        while (!(mQueue = mService.uploadQueue()).isEmpty()) {
            Log.d(TAG, "run: Starting pre-upload of" + mQueue.size() + "files.");
            JSONObject preUpload = doPreUpload();
            if (preUpload == null) {
                Log.w(TAG, "run: Preupload failed. ending UploadThread.");
                mService.onUploadThreadEnding();
                return;
            }

            Log.d(TAG, "run: Starting upload of" + mQueue.size() + "files.");
            if (!doUpload(preUpload)) {
                Log.w(TAG, "run: Upload failed, ending UploadThread.");
                mService.onUploadThreadEnding();
                return;
            }
            Log.d(TAG, "run: Did upload. Queue size is now " + mQueue.size() + "files.");
        }
        Log.d(TAG, "run: Queue empty; done.");
        mService.onUploadThreadEnding();
    }

    private boolean doUpload(JSONObject preUpload) {
        String uploadUrl = preUpload.optString("uploadUrl", "http://" + mHostPort + "/camli/upload");
        HttpPost uploadReq = new HttpPost(uploadUrl);
        UploadApiClient uploadApiClient = new UploadApiClient(mUA, mQueue, mService, uploadUrl);
        return uploadApiClient.doUpload(preUpload, uploadReq);
    }

    private JSONObject doPreUpload() {
        // Do the pre-upload
        HttpPost preReq = new HttpPost("http://" + mHostPort + "/camli/preupload");
        UploadApiClient uploadApiClient = new UploadApiClient(mUA, mQueue, mService);
        return uploadApiClient.doPreUpload(preReq);
    }
}
