package io.manun.camli;

import android.app.Service;
import android.content.Intent;
import android.content.SharedPreferences;
import android.os.IBinder;
import android.os.ParcelFileDescriptor;
import android.os.RemoteException;
import android.util.Log;

import androidx.annotation.Nullable;

import java.io.IOException;

public class UploadService extends Service {

    private static final String TAG = UploadService.class.getName();

    @Nullable
    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

//    private final IUploadService.Stub service = new IUploadService.Stub() {
//
//        private boolean mUploading = false;
//        private UploadThread mUploadThread = null;
//
//        @Override
//        public boolean addFile(ParcelFileDescriptor pfd) throws RemoteException {
//            SharedPreferences sp = getSharedPreferences(Preferences.NAME, 0);
//            HostPort hp = new HostPort(sp.getString(Preferences.HOST, ""));
//
//            if (!hp.isValid()) return false;
//            String password = sp.getString(Preferences.PASSWORD, "");
//
//            synchronized (this) {
//                if (!mUploading) {
//                    mUploading = true;
//                    mUploadThread = new UploadThread(hp, password);
//                    mUploadThread.start();
//                }
//            }
//
//            Log.d(TAG, "addFile for " + pfd + "; size=" + pfd.getStatSize());
//            try {
//                pfd.close();
//            } catch (IOException e) {
//                e.printStackTrace();
//            }
//            return true;
//        }
//
//        @Override
//        public void registerCallback(IStatusCallback ob) throws RemoteException {
//
//        }
//
//        @Override
//        public void unregisterCallback(IStatusCallback ob) throws RemoteException {
//
//        }
//
//        @Override
//        public boolean isUploading() throws RemoteException {
//            synchronized (this) {
//                return mUploading;
//            }
//        }
//
//        @Override
//        public void stop() throws RemoteException {
//
//        }
//
//        @Override
//        public void start() throws RemoteException {
//
//        }
//
//
//    };

}
