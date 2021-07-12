package io.manun.camli;


import android.os.ParcelFileDescriptor;
import android.os.SystemClock;
import android.util.Log;

import org.apache.http.Header;
import org.apache.http.HttpEntity;
import org.apache.http.message.BasicHeader;

import java.io.BufferedOutputStream;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.io.PrintWriter;
import java.util.ArrayList;
import java.util.LinkedList;
import java.util.List;
import java.util.concurrent.atomic.AtomicBoolean;

// TODO: test it
public class MultipartEntity implements HttpEntity {

    private static final String TAG = MultipartEntity.class.getName();

    private final List<QueuedFile> mFilesWritten = new ArrayList<>();

    private boolean mDone = false;
    private final String mBoundary;

    private final LinkedList<QueuedFile> mQueue;
    private final AtomicBoolean mStopRequested = new AtomicBoolean(false);
    private final UploadService mService;

    public MultipartEntity(LinkedList<QueuedFile> mQueue, UploadService mService) {
        mBoundary = "TODOLKSDJFLKSDJFLdslkjfjf23ojf0j30dm32LFDSJFLKSDJF";
        this.mQueue = mQueue;
        this.mService = mService;
    }

//    public MultipartEntity() {
//        // TODO: proper boundary
//        mBoundary = "TODOLKSDJFLKSDJFLdslkjfjf23ojf0j30dm32LFDSJFLKSDJF";
//    }

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
