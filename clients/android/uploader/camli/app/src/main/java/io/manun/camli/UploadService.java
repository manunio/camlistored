package io.manun.camli;

import android.app.Service;
import android.content.Intent;
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
        return service;
    }

    private final IUploadService.Stub service = new IUploadService.Stub() {
        @Override
        public void registerCallback(IStatusCallback ob) throws RemoteException {

        }

        @Override
        public void unregisterCallback(IStatusCallback ob) throws RemoteException {

        }

        @Override
        public boolean isUploading() throws RemoteException {
            return false;
        }

        @Override
        public void stop() throws RemoteException {

        }

        @Override
        public void start() throws RemoteException {

        }

        @Override
        public void addFile(ParcelFileDescriptor pfd) throws RemoteException {
            Log.d(TAG, "addFile for " + pfd + "; size=" + pfd.getStatSize());
            try {
                pfd.close();
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
    };

}
