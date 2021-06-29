package io.manun.camli;

import android.content.ComponentName;
import android.content.ContentResolver;
import android.content.Context;
import android.content.Intent;
import android.content.ServiceConnection;
import android.net.Uri;
import android.os.AsyncTask;
import android.os.Bundle;

import com.google.android.material.floatingactionbutton.FloatingActionButton;
import com.google.android.material.snackbar.Snackbar;

import androidx.appcompat.app.AppCompatActivity;
import androidx.appcompat.widget.Toolbar;
import androidx.core.app.NavUtils;

import android.os.Handler;
import android.os.IBinder;
import android.os.Looper;
import android.os.ParcelFileDescriptor;
import android.os.RemoteException;
import android.util.Log;
import android.view.View;

import android.view.Menu;
import android.view.MenuItem;

import java.io.FileDescriptor;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.net.URI;
import java.util.ArrayList;
import java.util.concurrent.Executor;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class MainActivity extends AppCompatActivity {

    private static final String TAG = MainActivity.class.getName();

    private IUploadService serviceStub = null;

    private final ArrayList<Uri> pendingUrisToUpload = new ArrayList<>();

    private final IStatusCallback.Stub statusCallback = new IStatusCallback.Stub() {
        @Override
        public void logToClient(String stuff) throws RemoteException {
            Log.d(TAG, "From service: " + stuff);
        }

        @Override
        public void onUploadStatusChange(boolean uploading) throws RemoteException {
            Log.d(TAG, "upload status change: " + uploading);
        }
    };

    private final ServiceConnection serviceConnection = new ServiceConnection() {
        @Override
        public void onServiceConnected(ComponentName name, IBinder service) {
            serviceStub = IUploadService.Stub.asInterface(service);
            Log.d(TAG, "Service connected");
            try {
                serviceStub.registerCallback(statusCallback);
                // Drain the queue from before the service was connected.
                for (Uri uri : pendingUrisToUpload) {
                    startDownloadOfUri(uri);
                }
                pendingUrisToUpload.clear();
            } catch (RemoteException e) {
                e.printStackTrace();
            }
        }

        @Override
        public void onServiceDisconnected(ComponentName name) {
            Log.d(TAG, "Service disconnected");
            serviceStub = null;
        }
    };

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        Toolbar toolbar = findViewById(R.id.toolbar);
        setSupportActionBar(toolbar);

        FloatingActionButton fab = findViewById(R.id.fab);
        fab.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                Snackbar.make(view, "Replace with your own action", Snackbar.LENGTH_LONG)
                        .setAction("Action", null).show();
            }
        });
    }

    @Override
    protected void onResume() {
        super.onResume();

        bindService(new Intent(this, UploadService.class), serviceConnection,
                Context.BIND_AUTO_CREATE);

        Intent intent = getIntent();
        String action = intent.getAction();
        Log.d(TAG, "onResume; action=" + action);
        if (Intent.ACTION_SEND.equals(action)) {
            handleSend(intent);
        } else if (Intent.ACTION_SEND_MULTIPLE.equals(action)) {
            handleSendMultiple(intent);
        }
    }

    @Override
    protected void onPause() {
        super.onPause();
        try {
            if (serviceStub != null) {
                serviceStub.unregisterCallback(statusCallback);
            }
        } catch (RemoteException e) {
            e.printStackTrace();
        }
        if (serviceConnection != null) {
            unbindService(serviceConnection);
        }
    }

    private void handleSendMultiple(Intent intent) {
//        TODO:
    }

    private void handleSend(Intent intent) {
        Bundle extras = intent.getExtras();
        if (extras == null) {
            Log.w(TAG, "expected extras in handleSend");
            return;
        }
        extras.keySet(); // unparcel
        Log.d(TAG, "handleSend; extras=" + extras);

        Object streamValue = extras.get("android.intent.extra.STREAM");
        if (!(streamValue instanceof Uri)) {
            Log.w(TAG, "Expected URI for STREAM; got: " + streamValue);
            return;
        }

        Uri uri = (Uri) streamValue;
        startDownloadOfUri(uri);
    }

    private void startDownloadOfUri(final Uri uri) {
        if (serviceStub == null) {
            Log.d(TAG, "serviceStub is null in startDownloadOfUri, enqueuing");
            pendingUrisToUpload.add(uri);
            return;
        }

        ExecutorService executor = Executors.newSingleThreadExecutor();
        Handler handler = new Handler(Looper.getMainLooper()); // only parameter-less constructor is deprecated.
        executor.execute(new Runnable() {
            @Override
            public void run() {
                // Background work here.
                try {
                    serviceStub.enqueueUpload(uri);
                } catch (RemoteException e) {
                    Log.d(TAG, "failure to enqueue upload", e);
                }

                handler.post(new Runnable() {
                    @Override
                    public void run() {
                        // UI thread work here.
                        Log.d(TAG, "UI thread");
                    }
                });
            }
        });
    }

    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        // Inflate the menu; this adds items to the action bar if it is present.
        getMenuInflater().inflate(R.menu.menu_main, menu);
        return true;
    }

    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        // Handle action bar item clicks here. The action bar will
        // automatically handle clicks on the Home/Up button, so long
        // as you specify a parent activity in AndroidManifest.xml.
        int id = item.getItemId();
        //noinspection SimplifiableIfStatement
        if (id == R.id.action_settings) {
            SettingsActivity.show(this);
            return true;
        }

        return super.onOptionsItemSelected(item);
    }
}
