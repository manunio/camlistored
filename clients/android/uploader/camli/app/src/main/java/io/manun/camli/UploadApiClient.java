package io.manun.camli;

import android.util.Log;

import org.apache.http.HttpResponse;
import org.apache.http.StatusLine;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.impl.client.DefaultHttpClient;
import org.apache.http.message.BasicNameValuePair;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.IOException;
import java.io.UnsupportedEncodingException;
import java.util.ArrayList;
import java.util.LinkedList;
import java.util.List;

public class UploadApiClient {
    private static final String TAG = UploadThread.class.getName();
    private final DefaultHttpClient mUA;
    private final LinkedList<QueuedFile> mQueue;
    private final UploadService mService;
    private String uploadUrl;

    public UploadApiClient(DefaultHttpClient mUA,
                           LinkedList<QueuedFile> mQueue,
                           UploadService mService,
                           String uploadUrl) {
        this.mUA = mUA;
        this.mQueue = mQueue;
        this.mService = mService;
        this.uploadUrl = uploadUrl;
    }


    public UploadApiClient(DefaultHttpClient mUA,
                           LinkedList<QueuedFile> mQueue,
                           UploadService mService
    ) {
        this.mUA = mUA;
        this.mQueue = mQueue;
        this.mService = mService;
    }

    public boolean doUpload(JSONObject preUpload, HttpPost uploadReq) {
        Log.d(TAG, "JSON: " + preUpload);
        Log.d(TAG, "uploadURL is: " + uploadUrl);
        MultipartEntity entity = new MultipartEntity(mQueue, mService);
        uploadReq.setEntity(entity);
        HttpResponse uploadRes = null;
        try {
            uploadRes = mUA.execute(uploadReq);
        } catch (IOException e) {
            Log.e(TAG, "run: upload error", e);
            return false;
        }
        Log.d(TAG, "doUpload: response: " + uploadRes);
        StatusLine statusLine = uploadRes.getStatusLine();
        Log.d(TAG, "doUpload: response code: " + statusLine);
        // TODO: check response body, once response body is defined?
        if (statusLine == null || statusLine.getStatusCode() < 200 ||
                statusLine.getStatusCode() > 299) {
            Log.d(TAG, "doUpload: upload error.");
            // TODO: back-off? or probably in the service layer.
            return false;
        }
        for (QueuedFile qf : entity.getFilesWritten()) {
            // TODO: only do this if acknowledge in JSON response?
            Log.d(TAG, "doUpload: upload complete for: " + qf);
            mService.onUploadComplete(qf);
        }
        Log.d(TAG, "doUpload: returning true.");
        return true;
    }


    public JSONObject doPreUpload(HttpPost preReq) {
        // Do the pre-upload

        List<BasicNameValuePair> uploadKeys = new ArrayList<>();
        uploadKeys.add(new BasicNameValuePair("camliversion", "1"));

        int n = 0;
        for (QueuedFile qf : mQueue) {
            uploadKeys.add(new BasicNameValuePair("blob" + (++n), qf.getContentName()));
        }

        try {
            preReq.setEntity(new UrlEncodedFormEntity(uploadKeys));
        } catch (UnsupportedEncodingException e) {
            Log.e(TAG, "error", e);
            return null;
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
            return null;
        } catch (JSONException e) {
            Log.e(TAG, "preupload JSON parse error from: " + jsonSlurp, e);
            return null;
        }
        return preUpload;
    }

}
