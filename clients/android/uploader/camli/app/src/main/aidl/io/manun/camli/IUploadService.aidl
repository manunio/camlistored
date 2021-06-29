// IUploadService.aidl
package io.manun.camli;

import io.manun.camli.IStatusCallback;
import android.net.Uri;

// Declare any non-default types here with import statements

interface IUploadService {
  void registerCallback(IStatusCallback ob);
  void unregisterCallback(IStatusCallback ob);

  int queueSize();
  boolean isUploading();

  // Returns true if thread was running and we requested it to be stopped.
  boolean pause();

  // Returns true if upload wasn't already in progress and new upload
  // thread was started.
  boolean resume();

  // Enqueues a new file to be uploaded (a file:// or content:// URL). Does disk I/O,
  // so should be called from an AsyncTask(old) / Executor(new).
  // Returns false if server not configured.
  boolean enqueueUpload(in Uri uri);
}