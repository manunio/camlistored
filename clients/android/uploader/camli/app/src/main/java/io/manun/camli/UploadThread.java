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

    private final HostPort mHostPort;
    private final String mPassword;

    private final AtomicBoolean mStopRequested = new AtomicBoolean(false);

    public UploadThread(HostPort mHostPort, String mPassword) {
        this.mHostPort = mHostPort;
        this.mPassword = mPassword;
    }

    private void stopPlease() {
        mStopRequested.set(false);
    }

    @Override
    public void run() {
        if (!mHostPort.isValid())
            return;
        Log.d(TAG, "Running UploadThread for " + mHostPort);

        DefaultHttpClient ua = new DefaultHttpClient();
        HttpRequestFactory reqFactory = new DefaultHttpRequestFactory();

        CredentialsProvider creds = new BasicCredentialsProvider();
        creds.setCredentials(AuthScope.ANY,
                new UsernamePasswordCredentials("TODO-DUMMY-USER", mPassword)
        );
        ua.setCredentialsProvider(creds);
        // DO the pre-upload
        HttpPost preReq = new HttpPost("http://" + mHostPort +
                "/camli/preupload");

        List<BasicNameValuePair> uploadKeys = new ArrayList<>();
        uploadKeys.add(new BasicNameValuePair("camliversion", "1"));
        try {
            preReq.setEntity(new UrlEncodedFormEntity(uploadKeys));
        } catch (UnsupportedEncodingException e) {
            Log.e(TAG, "error", e);
            return;
        }
        try {
            HttpResponse res = ua.execute(preReq);
            Log.d(TAG, "response: " + res);
            Log.d(TAG, "response code: " + res.getStatusLine());
            Log.d(TAG, "entity: " + res.getEntity());

            String jsonSlurp = Util.slurp(res.getEntity().getContent());
            Log.d(TAG, "JSON content: " + jsonSlurp);
            JSONObject json = new JSONObject(jsonSlurp);
            Log.d(TAG, "JSON response: " + json);
        } catch (IOException e) {
            Log.e(TAG, "preupload error", e);
        } catch (JSONException e) {
            Log.e(TAG, "JSON parse error", e);
        }
    }
}
