package io.manun.camli;

import android.util.Log;

import org.apache.http.HttpRequestFactory;
import org.apache.http.HttpResponse;
import org.apache.http.auth.AuthScope;
import org.apache.http.auth.UsernamePasswordCredentials;
import org.apache.http.client.ClientProtocolException;
import org.apache.http.client.CredentialsProvider;
import org.apache.http.client.HttpClient;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.impl.DefaultHttpRequestFactory;
import org.apache.http.impl.client.BasicCredentialsProvider;
import org.apache.http.impl.client.DefaultHttpClient;
import org.apache.http.message.BasicNameValuePair;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.UnsupportedEncodingException;
import java.net.HttpURLConnection;
import java.net.MalformedURLException;
import java.net.ProtocolException;
import java.net.URL;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.atomic.AtomicBoolean;

public class UploadThread extends Thread {
    private static final String TAG = UploadThread.class.getName();

    private final UploadService mService;
    private final HostPort mHostPort;

    private final AtomicBoolean mStopRequested = new AtomicBoolean(false);

    private final DefaultHttpClient mUA = new DefaultHttpClient();

    public UploadThread(UploadService mService, HostPort mHostPort, String mPassword) {
        this.mService = mService;
        this.mHostPort = mHostPort;

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

        List<QueuedFile> queue = mService.uploadQueue();
        if (queue.isEmpty()) {
            Log.d(TAG, "Queue empty; done");
            return;
        }

        // DO the pre-upload
        HttpPost preReq = new HttpPost("http://" + mHostPort +
                "/camli/preupload");

        List<BasicNameValuePair> uploadKeys = new ArrayList<>();
        uploadKeys.add(new BasicNameValuePair("camliversion", "1"));

        int n = 0;
        for (QueuedFile qf : queue) {
            uploadKeys.add(new BasicNameValuePair("blob" + (++n), qf.getContentName()));
        }

        try {
            preReq.setEntity(new UrlEncodedFormEntity(uploadKeys));
        } catch (UnsupportedEncodingException e) {
            Log.e(TAG, "error", e);
            return;
        }
        JSONObject preUpload = null;
        String jsonSlurp = null;
        try {
            HttpResponse res = mUA.execute(preReq);
            Log.d(TAG, "response: " + res);
            Log.d(TAG, "response code: " + res.getStatusLine());
            // TODO: check response code
            jsonSlurp = Util.slurp(res.getEntity().getContent());
            preUpload = new JSONObject(jsonSlurp);
        } catch (IOException e) {
            Log.e(TAG, "preupload error", e);
            return;
        } catch (JSONException e) {
            Log.e(TAG, "preupload JSON parse error from: " + jsonSlurp, e);
            return;
        }

        Log.d(TAG, "JSON: " + preUpload);
        String uploadUrl = preUpload
                .optString("uploadUrl", "http://" + mHostPort + "/camli/upload");
        Log.d(TAG, "uploadURL is: " + uploadUrl);
        HttpPost uploadReq = new HttpPost(uploadUrl);

    }
}
