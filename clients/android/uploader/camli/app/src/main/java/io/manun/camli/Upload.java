package io.manun.camli;

import android.os.ParcelFileDescriptor;
import android.os.SystemClock;
import android.util.Log;

import org.apache.http.Header;
import org.apache.http.HttpEntity;
import org.apache.http.HttpResponse;
import org.apache.http.StatusLine;
import org.apache.http.auth.AuthScope;
import org.apache.http.auth.UsernamePasswordCredentials;
import org.apache.http.client.CredentialsProvider;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.impl.client.BasicCredentialsProvider;
import org.apache.http.impl.client.DefaultHttpClient;
import org.apache.http.message.BasicHeader;
import org.apache.http.message.BasicNameValuePair;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.BufferedOutputStream;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.io.PrintWriter;
import java.io.UnsupportedEncodingException;
import java.util.ArrayList;
import java.util.LinkedList;
import java.util.List;
import java.util.concurrent.atomic.AtomicBoolean;

public class Upload {
    private static final String TAG = Upload.class.getName();

    private final UploadService mService;
    private final HostPort mHostPort;
    private LinkedList<QueuedFile> mQueue;

    private final AtomicBoolean mStopRequested = new AtomicBoolean(false);

    private final DefaultHttpClient mUA = new DefaultHttpClient();

    public Upload(UploadService mService, HostPort mHostPort, String mPassword) {
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
        Log.d(TAG, "JSON: " + preUpload);
        String uploadUrl = preUpload
                .optString("uploadUrl", "http://" + mHostPort + "/camli/upload");
        Log.d(TAG, "uploadURL is: " + uploadUrl);
        HttpPost uploadReq = new HttpPost(uploadUrl);
        MultipartEntity entity = new MultipartEntity();
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

    private JSONObject doPreUpload() {
        // Do the pre-upload
        HttpPost preReq = new HttpPost("http://" + mHostPort +
                "/camli/preupload");

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

    private class MultipartEntity implements HttpEntity {

        private final List<QueuedFile> mFilesWritten = new ArrayList<>();

        private boolean mDone = false;
        private final String mBoundary;

        public MultipartEntity() {
            // TODO: proper boundary
            mBoundary = "TODOLKSDJFLKSDJFLdslkjfjf23ojf0j30dm32LFDSJFLKSDJF";
        }

        @Override
        public boolean isRepeatable() {
            Log.d(TAG, "isRepeatable: ");
            // Well, not really, but needs to be for DefaultRequestDirector
            return true;
        }

        @Override
        public boolean isChunked() {
            Log.d(TAG, "isChunked: ");
            return false;
        }

        @Override
        public long getContentLength() {
            return -1; //unknown
        }

        @Override
        public Header getContentType() {
            return new BasicHeader("Content-Type", "multipart/form-data; boundary=" + mBoundary);
        }

        @Override
        public Header getContentEncoding() {
            return null; // unknown
        }

        @Override
        public InputStream getContent() throws IOException, UnsupportedOperationException {
            Log.d(TAG, "getContent: ");
            throw new RuntimeException("unexpected getContent call");
        }


        @Override
        public boolean isStreaming() {
            Log.d(TAG, "isStreaming: ");
            return !mDone;
        }

        @Override
        public void consumeContent() throws IOException {
            // From the docs: "The name of this method is misnomer ...
            // This method is called to indicate that the content of this entity
            // is no longer required. All entity implementations are expected to
            // release all allocated resources as a result of this method
            // invocation."
            Log.d(TAG, "consumeContent: ");
            mDone = true;
        }

        @Override
        public void writeTo(OutputStream out) throws IOException {
            BufferedOutputStream bos = new BufferedOutputStream(out);
            PrintWriter pw = new PrintWriter(bos);
            byte[] buf = new byte[1024];

            int bytesWritten = 0;
            long timeStarted = SystemClock.uptimeMillis();

            for (QueuedFile qf : mQueue) {
                ParcelFileDescriptor pfd = mService.getFileDescriptor(qf.getUri());
                if (pfd == null) {
                    // TODO: report some errors up to user?
                    mQueue.removeFirst();
                    continue;
                }
                startNewBoundary(pw);
                pw.flush();
                pw.print("Content-Disposition: form-data; name=");
                pw.print(qf.getContentName());
                pw.print("\r\n\r\n");
                pw.flush();

                FileInputStream fis = new FileInputStream(pfd.getFileDescriptor());
                int n;
                while ((n = fis.read(buf)) != -1) {
                    bos.write(buf, 0, n);
                    if (mStopRequested.get()) {
                        Log.d(TAG, "writeTo: Stopping upload pre-maturely");
                        pfd.close();
                        return;
                    }
                }
                bos.flush();
                pfd.close();
                mQueue.removeFirst();
                // TODO: notification of update
                Log.d(TAG, "writeTo: write of " + qf.getContentName() + " complete.");
                mFilesWritten.add(qf);

                if (bytesWritten > 1024 * 1024) {
                    Log.d(TAG, "writeTo: enough bytes written, stopping writing after " + bytesWritten);
                    // stop after 1MB to get response.
                    // TODO: make this smarter, configurable, time-based.
                    break;
                }
            }
            endBoundary(pw);
            pw.flush();
            Log.d(TAG, "writeTo: finished writing upload MIME body.");
        }

        private void startNewBoundary(PrintWriter pw) {
            pw.print("\r\n--");
            pw.print(mBoundary);
            pw.print("\r\n");
        }

        private void endBoundary(PrintWriter pw) {
            pw.print("\r\n--");
            pw.print(mBoundary);
            pw.print("--\r\n");
        }

        public List<QueuedFile> getFilesWritten() {
            return mFilesWritten;
        }
    }

}
