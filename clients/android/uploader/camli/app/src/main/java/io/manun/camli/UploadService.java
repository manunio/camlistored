package io.manun.camli;

import android.app.Service;
import android.content.ContentResolver;
import android.content.Intent;
import android.content.SharedPreferences;
import android.net.Uri;
import android.os.Binder;
import android.os.IBinder;
import android.os.ParcelFileDescriptor;
import android.os.RemoteException;
import android.util.Log;

import androidx.annotation.Nullable;

import org.apache.http.client.methods.HttpPost;

import java.io.FileNotFoundException;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.LinkedList;
import java.util.List;
import java.util.Set;

public class UploadService extends Service {

    private static final String TAG = UploadService.class.getName();

    // Guarded by 'this':
    private boolean mUploading = false;
    private UploadThread mUploadThread = null;
    private final Set<QueuedFile> mQueueSet = new HashSet<>();
    private final LinkedList<QueuedFile> mQueueList = new LinkedList<>();


    public class UploadBinder extends Binder {
        UploadService getService() {
            return UploadService.this;
        }
    }

    @Nullable
    @Override
    public IBinder onBind(Intent intent) {
        return service;
    }

    // Called by UploadThread to get stuff to do, Caller owns returned list.
    LinkedList<QueuedFile> uploadQueue() {
        synchronized (this) {
            return new LinkedList<>(mQueueList);
        }
    }

    void onUploadThreadEnding() {
        synchronized (this) {
            mUploadThread = null;
            mUploading = false;
        }
    }

    void onUploadComplete(QueuedFile qf) {
        synchronized (this) {
            boolean removedSet = mQueueSet.remove(qf);
            boolean removedList = mQueueList.remove(qf); //TODO: ghetto, linter
            // scan
            Log.d(TAG, "onUploadComplete: removing of" + qf + "; removedSet="
                    + removedSet + "; removedList=" + removedList);
        }
    }

    private final IUploadService.Stub service = new IUploadService.Stub() {
        @Override
        public boolean enqueueUpload(Uri uri) throws RemoteException {
            SharedPreferences sp = getSharedPreferences(Preferences.NAME, 0);
            HostPort hp = new HostPort(sp.getString(Preferences.HOST, ""));

            if (!hp.isValid()) return false;

            ParcelFileDescriptor pfd = getFileDescriptor(uri);
            if (pfd == null) {
                return false;
            }

            String sha1 = Util.getSha1(pfd.getFileDescriptor());
            Log.d(TAG, "sha1 of file is: " + sha1);
            Log.d(TAG, "size of file is: " + pfd.getStatSize());
            QueuedFile qf = new QueuedFile(sha1, uri);

            synchronized (UploadService.this) {
                if (mQueueSet.contains(qf)) {
                    return false;
                }
                mQueueSet.add(qf);
                mQueueList.add(qf);
                if (!mUploading) {
                    resume();
                }
            }
            return true;
        }

        @Override
        public void registerCallback(IStatusCallback ob) throws RemoteException {

        }

        @Override
        public void unregisterCallback(IStatusCallback ob) throws RemoteException {

        }

        @Override
        public int queueSize() throws RemoteException {
            synchronized (UploadService.this) {
                return mQueueList.size();
            }
        }

        @Override
        public boolean isUploading() throws RemoteException {
            synchronized (UploadService.this) {
                return mUploading;
            }
        }

        @Override
        public boolean pause() throws RemoteException {
            synchronized (UploadService.this) {
                if (mUploadThread != null) {
                    mUploadThread.stopPlease();
                    return true;
                }
                return false;
            }
        }

        @Override
        public boolean resume() throws RemoteException {
            synchronized (UploadService.this) {
                if (mUploadThread != null) {
                    return false;
                }
                mUploading = true;
                SharedPreferences sp = getSharedPreferences(Preferences.NAME, 0);
                HostPort hp = new HostPort(sp.getString(Preferences.HOST, ""));
                if (!hp.isValid()) {
                    return false;
                }
                String password = sp.getString(Preferences.PASSWORD, "");
                mUploadThread = new UploadThread(UploadService.this, hp, password);
                mUploadThread.start();
                return true;
            }
        }
    };

    public ParcelFileDescriptor getFileDescriptor(Uri uri) {
        ContentResolver cr = getContentResolver();
        try {
            return cr.openFileDescriptor(uri, "r");
        } catch (FileNotFoundException e) {
            Log.e(TAG, "getFileDescriptor: FileNotFound for" + uri, e);
            return null;
        }
    }
}
